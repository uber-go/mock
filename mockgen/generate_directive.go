package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func generateDirective(dst string, flags *flag.FlagSet, argv []string) (string, error) {
	args := slices.Clone(argv)
	if dst != "" && filepath.Dir(dst) != "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory for go:generate directive: %w", err)
		}
		dir := filepath.Dir(dst)
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(wd, dir)
		}
		if dir != wd {
			args = rebaseGenerateArgs(wd, dir, flags, argv[0])
		}
	}

	for i, arg := range args {
		// go generate expands environment variables even in quoted arguments.
		arg = strings.ReplaceAll(arg, "$", "${DOLLAR}")
		if arg == "" || strings.HasPrefix(arg, `"`) || strings.ContainsAny(arg, " \t\r\n") {
			arg = strconv.Quote(arg)
		}
		args[i] = arg
	}
	return strings.Join(args, " "), nil
}

func rebaseGenerateArgs(wd, dir string, flags *flag.FlagSet, program string) []string {
	rebase := func(name string) string {
		if name == "" || filepath.IsAbs(name) {
			return name
		}
		absolute := filepath.Join(wd, name)
		if relative, err := filepath.Rel(dir, absolute); err == nil {
			return relative
		}
		return absolute
	}
	if filepath.Base(program) != program && !filepath.IsAbs(program) {
		program = rebase(program)
		if filepath.Base(program) == program {
			program = "." + string(filepath.Separator) + program
		}
	}

	// Use the parsed flags so both -flag=value and -flag value work without
	// reimplementing flag parsing.
	args := []string{program}
	flags.Visit(func(f *flag.Flag) {
		value := f.Value.String()
		switch f.Name {
		case "archive", "copyright_file", "destination", "model_gob", "source":
			value = rebase(value)
		case "aux_files":
			files := strings.Split(value, ",")
			for i, file := range files {
				if pkg, name, ok := strings.Cut(file, "="); ok {
					files[i] = pkg + "=" + rebase(name)
				}
			}
			value = strings.Join(files, ",")
		}
		args = append(args, "-"+f.Name+"="+value)
	})

	positional := slices.Clone(flags.Args())
	if *archive == "" && *modelGob == "" && *source == "" && len(positional) > 0 {
		pkg := filepath.ToSlash(positional[0])
		if pkg == "." || pkg == ".." || strings.HasPrefix(pkg, "./") || strings.HasPrefix(pkg, "../") {
			pkg = filepath.ToSlash(rebase(positional[0]))
			if pkg != "." && pkg != ".." && !strings.HasPrefix(pkg, "../") && !filepath.IsAbs(pkg) {
				pkg = "./" + pkg
			}
			positional[0] = pkg
		}
	}
	if len(positional) > 0 {
		args = append(args, "--")
	}
	return append(args, positional...)
}
