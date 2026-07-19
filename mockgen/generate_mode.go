package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/mock/mockgen/model"
)

type generateDirectiveEnv struct {
	GOFILE    string
	GOLINE    string
	GOPACKAGE string
	DOLLAR    string
}

type generateTargetDirective struct {
	dir  string
	args []string
}

type generateTargetOptions struct {
	dir                    string
	archive                string
	source                 string
	destination            string
	mockNames              string
	packageOut             string
	selfPackage            string
	writeCmdComment        bool
	writePkgComment        bool
	writeSourceComment     bool
	writeGenerateDirective bool
	copyrightFile          string
	buildConstraint        string
	typed                  bool
	imports                string
	auxFiles               string
	modelGob               string
	excludeInterfaces      string
	debugParser            bool
	positionals            []string
	originalArgs           []string
	buildFlags             string
}

type executableGenerateTarget struct {
	options     generateTargetOptions
	pkg         *model.Package
	packageName string
}

type listedGeneratePackage struct {
	Dir         string
	Name        string
	GoFiles     []string
	TestGoFiles []string
}

var packageNameCache map[string]string

func runGenerateMode(patterns []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	directives, err := discoverGenerateTargets(wd, patterns)
	if err != nil {
		return err
	}

	executable := make([]executableGenerateTarget, 0, len(directives))
	for _, directive := range directives {
		options, err := parseGenerateTargetOptions(directive)
		if err != nil {
			return err
		}
		target, err := loadGenerateTarget(options)
		if err != nil {
			return err
		}
		executable = append(executable, target)
	}

	primePackageNameCache(executable)
	for i := range executable {
		if err := executeGenerateTarget(executable[i]); err != nil {
			return err
		}
	}
	return nil
}

func discoverGenerateTargets(root string, patterns []string) ([]generateTargetDirective, error) {
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	args := append([]string{"list", "-json=Dir,Name,GoFiles,TestGoFiles"}, patterns...)
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list generate packages: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(out))
	var targets []generateTargetDirective
	for dec.More() {
		var pkg listedGeneratePackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("decode go list output: %w", err)
		}
		pkgTargets, err := discoverGenerateTargetsInPackage(pkg)
		if err != nil {
			return nil, err
		}
		targets = append(targets, pkgTargets...)
	}
	return targets, nil
}

func discoverGenerateTargetsInPackage(pkg listedGeneratePackage) ([]generateTargetDirective, error) {
	files := append([]string{}, pkg.GoFiles...)
	files = append(files, pkg.TestGoFiles...)
	var targets []generateTargetDirective
	for _, name := range files {
		path := filepath.Join(pkg.Dir, name)
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}
		scanner := bufio.NewScanner(f)
		line := 0
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if !strings.HasPrefix(text, "//go:generate") {
				continue
			}
			target, ok, err := parseGenerateDirective(text, generateDirectiveEnv{
				GOFILE:    name,
				GOLINE:    strconv.Itoa(line),
				GOPACKAGE: pkg.Name,
				DOLLAR:    "$",
			})
			if err != nil {
				_ = f.Close()
				return nil, fmt.Errorf("parse %s:%d: %w", path, line, err)
			}
			if ok {
				target.dir = pkg.Dir
				targets = append(targets, target)
			}
		}
		if err := scanner.Err(); err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("scan %s: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("close %s: %w", path, err)
		}
	}
	return targets, nil
}

func parseGenerateTargetOptions(directive generateTargetDirective) (generateTargetOptions, error) {
	options := generateTargetOptions{
		dir:                    directive.dir,
		writeCmdComment:        true,
		writePkgComment:        true,
		writeSourceComment:     true,
		writeGenerateDirective: false,
	}
	fs := flag.NewFlagSet("mockgen generate directive", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&options.archive, "archive", "", "")
	fs.StringVar(&options.source, "source", "", "")
	fs.StringVar(&options.destination, "destination", "", "")
	fs.StringVar(&options.mockNames, "mock_names", "", "")
	fs.StringVar(&options.packageOut, "package", "", "")
	fs.StringVar(&options.selfPackage, "self_package", "", "")
	fs.BoolVar(&options.writeCmdComment, "write_command_comment", true, "")
	fs.BoolVar(&options.writePkgComment, "write_package_comment", true, "")
	fs.BoolVar(&options.writeSourceComment, "write_source_comment", true, "")
	fs.BoolVar(&options.writeGenerateDirective, "write_generate_directive", false, "")
	fs.StringVar(&options.copyrightFile, "copyright_file", "", "")
	fs.StringVar(&options.buildConstraint, "build_constraint", "", "")
	fs.StringVar(&options.buildFlags, "build_flags", "", "")
	fs.BoolVar(&options.typed, "typed", false, "")
	fs.StringVar(&options.imports, "imports", "", "")
	fs.StringVar(&options.auxFiles, "aux_files", "", "")
	fs.StringVar(&options.modelGob, "model_gob", "", "")
	fs.StringVar(&options.excludeInterfaces, "exclude_interfaces", "", "")
	fs.BoolVar(&options.debugParser, "debug_parser", false, "")
	if err := fs.Parse(directive.args); err != nil {
		return generateTargetOptions{}, err
	}
	options.positionals = fs.Args()
	options.originalArgs = append([]string(nil), directive.args...)
	return options, nil
}

func loadGenerateTarget(options generateTargetOptions) (executableGenerateTarget, error) {
	var pkg *model.Package
	var err error
	var packageName string
	err = withWorkingDir(options.dir, func() error {
		switch {
		case options.modelGob != "":
			pkg, err = gobMode(options.modelGob)
		case options.source != "":
			if len(options.positionals) > 1 {
				return errors.New("-source mode accepts at most one argument")
			}
			var includeNames []string
			if len(options.positionals) == 1 {
				includeNames = strings.Split(options.positionals[0], ",")
			}
			pkg, err = sourceModeWithOptions(options.source, options.imports, options.auxFiles, options.excludeInterfaces, includeNames)
		case options.archive != "":
			if len(options.positionals) < 1 {
				return errors.New("archive mode expects at least one argument")
			}
			packageName = options.positionals[0]
			var interfaces []string
			if len(options.positionals) > 1 {
				interfaces = strings.Split(options.positionals[1], ",")
			}
			pkg, err = parseExportFile(packageName, interfaces, options.archive)
		default:
			if len(options.positionals) != 2 {
				return errors.New("package mode expects exactly two arguments")
			}
			packageName = options.positionals[0]
			interfaces := strings.Split(options.positionals[1], ",")
			if packageName == "." {
				packageName, err = packageNameOfDir(options.dir)
				if err != nil {
					return err
				}
			}
			oldBuildFlags := *buildFlags
			*buildFlags = options.buildFlags
			defer func() { *buildFlags = oldBuildFlags }()
			parser := packageModeParser{}
			pkg, err = parser.parsePackage(packageName, interfaces)
		}
		return err
	})
	if err != nil {
		return executableGenerateTarget{}, err
	}
	return executableGenerateTarget{options: options, pkg: pkg, packageName: packageName}, nil
}

func primePackageNameCache(targets []executableGenerateTarget) {
	paths := make(map[string]bool)
	for _, target := range targets {
		for importPath := range target.pkg.Imports() {
			paths[importPath] = true
		}
		paths[gomockImportPath] = true
		for _, intf := range target.pkg.Interfaces {
			if len(intf.Methods) > 0 {
				paths["reflect"] = true
				break
			}
		}
	}
	all := make([]string, 0, len(paths))
	for importPath := range paths {
		all = append(all, importPath)
	}
	sort.Strings(all)
	packageNameCache = make(map[string]string, len(all))
	_ = createPackageMap(all)
}

func executeGenerateTarget(target executableGenerateTarget) error {
	return withWorkingDir(target.options.dir, func() error {
		return writeGenerateTarget(target)
	})
}

func writeGenerateTarget(target executableGenerateTarget) error {
	options := target.options
	outputPackageName := options.packageOut
	if outputPackageName == "" {
		outputPackageName = "mock_" + sanitize(target.pkg.Name)
	}

	outputPackagePath := options.selfPackage
	if outputPackagePath == "" && options.destination != "" {
		dstPath, err := filepath.Abs(filepath.Dir(options.destination))
		if err == nil {
			if pkgPath, err := parsePackageImport(dstPath); err == nil {
				outputPackagePath = pkgPath
			}
		}
	}

	g := &generator{buildConstraint: options.buildConstraint}
	if options.source != "" {
		g.filename = options.source
	} else if options.archive != "" {
		g.filename = options.archive
	} else {
		g.srcPackage = target.packageName
		if len(options.positionals) > 1 {
			g.srcInterfaces = options.positionals[1]
		}
	}
	g.destination = options.destination
	if options.mockNames != "" {
		g.mockNames = parseMockNames(options.mockNames)
	}
	if options.copyrightFile != "" {
		header, err := os.ReadFile(options.copyrightFile)
		if err != nil {
			return fmt.Errorf("read copyright file: %w", err)
		}
		g.copyrightHeader = string(header)
	}

	oldArgs := os.Args
	os.Args = append([]string{"mockgen"}, options.originalArgs...)
	oldImports, oldSelfPackage, oldTyped := *imports, *selfPackage, *typed
	oldWriteCmd, oldWritePkg, oldWriteSource := *writeCmdComment, *writePkgComment, *writeSourceComment
	oldWriteGenerate := *writeGenerateDirective
	*imports, *selfPackage, *typed = options.imports, options.selfPackage, options.typed
	*writeCmdComment, *writePkgComment, *writeSourceComment = options.writeCmdComment, options.writePkgComment, options.writeSourceComment
	*writeGenerateDirective = options.writeGenerateDirective
	defer func() {
		os.Args = oldArgs
		*imports, *selfPackage, *typed = oldImports, oldSelfPackage, oldTyped
		*writeCmdComment, *writePkgComment, *writeSourceComment = oldWriteCmd, oldWritePkg, oldWriteSource
		*writeGenerateDirective = oldWriteGenerate
	}()

	if err := g.Generate(target.pkg, outputPackageName, outputPackagePath); err != nil {
		return err
	}
	output := g.Output()
	if options.destination == "" {
		_, err := os.Stdout.Write(output)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(options.destination), os.ModePerm); err != nil {
		return err
	}
	existing, err := os.ReadFile(options.destination)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(existing) == len(output) && bytes.Equal(existing, output) {
		return nil
	}
	return os.WriteFile(options.destination, output, 0o666)
}

func withWorkingDir(dir string, fn func() error) error {
	old, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer func() { _ = os.Chdir(old) }()
	return fn()
}

func parseGenerateDirective(line string, env generateDirectiveEnv) (generateTargetDirective, bool, error) {
	const prefix = "//go:generate"
	if !strings.HasPrefix(line, prefix) {
		return generateTargetDirective{}, false, nil
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if rest == "" {
		return generateTargetDirective{}, false, nil
	}
	fields, err := splitGenerateDirective(rest)
	if err != nil {
		return generateTargetDirective{}, false, err
	}
	if len(fields) == 0 {
		return generateTargetDirective{}, false, nil
	}
	if !isMockgenInvocation(fields) {
		return generateTargetDirective{}, false, nil
	}
	mockgenArgs := mockgenInvocationArgs(fields)
	for i := range mockgenArgs {
		mockgenArgs[i] = expandGenerateEnv(mockgenArgs[i], env)
	}
	return generateTargetDirective{args: mockgenArgs}, true, nil
}

func isMockgenInvocation(fields []string) bool {
	if len(fields) == 0 {
		return false
	}
	if commandBase(fields[0]) == "mockgen" {
		return true
	}
	return len(fields) >= 3 && fields[0] == "go" && fields[1] == "run" && strings.Contains(fields[2], "mockgen")
}

func mockgenInvocationArgs(fields []string) []string {
	if commandBase(fields[0]) == "mockgen" {
		return append([]string(nil), fields[1:]...)
	}
	return append([]string(nil), fields[3:]...)
}

func commandBase(cmd string) string {
	base := filepath.Base(cmd)
	if runtime.GOOS == "windows" {
		base = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return base
}

func expandGenerateEnv(s string, env generateDirectiveEnv) string {
	vars := map[string]string{
		"GOFILE":    env.GOFILE,
		"GOLINE":    env.GOLINE,
		"GOPACKAGE": env.GOPACKAGE,
		"DOLLAR":    env.DOLLAR,
	}
	return os.Expand(s, func(name string) string {
		if v, ok := vars[name]; ok {
			return v
		}
		return os.Getenv(name)
	})
}

func splitGenerateDirective(s string) ([]string, error) {
	var fields []string
	for len(s) > 0 {
		s = strings.TrimLeft(s, " \t")
		if s == "" {
			break
		}
		field, rest, err := nextGenerateField(s)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
		s = rest
	}
	return fields, nil
}

func nextGenerateField(s string) (field, rest string, err error) {
	var b strings.Builder
	quote := rune(0)
	escaped := false
	for i, r := range s {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if quote != '\'' && r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				b.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"', '`':
			quote = r
		case ' ', '\t':
			return b.String(), s[i+1:], nil
		default:
			b.WriteRune(r)
		}
	}
	if escaped {
		return "", "", fmt.Errorf("unfinished escape")
	}
	if quote != 0 {
		return "", "", fmt.Errorf("unterminated quote")
	}
	return b.String(), "", nil
}
