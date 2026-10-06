package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestFileParser_ParseFile(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
	}

	pkg, err := p.parseFile("", file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	checkGreeterImports(t, p.imports)

	expectedName := "greeter"
	if pkg.Name != expectedName {
		t.Fatalf("Expected name to be %v but got %v", expectedName, pkg.Name)
	}

	expectedInterfaceName := "InputMaker"
	if pkg.Interfaces[0].Name != expectedInterfaceName {
		t.Fatalf("Expected interface name to be %v but got %v", expectedInterfaceName, pkg.Interfaces[0].Name)
	}
}

func TestFileParser_ParsePackage(t *testing.T) {
	fs := token.NewFileSet()
	_, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		packages:           make(map[string]*fileParser),
		importedInterfaces: newInterfaceCache(),
	}

	path := "go.uber.org/mock/mockgen/internal/tests/custom_package_name/greeter"
	newP, err := p.parsePackage(path)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	checkGreeterImports(t, newP.importedInterfaces.Get(path, "InputMaker").fileImports())
}

// faux.go and conflict.go import different packages named log.
func TestFileParser_ParsePackage_FileScopedImports(t *testing.T) {
	const base = "go.uber.org/mock/mockgen/internal/tests/import_embedded_interface"
	path := base + "/faux"
	p := fileParser{packages: make(map[string]*fileParser)}
	newP, err := p.parsePackage(path)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	iface, err := newP.parseInterface("Foreign", path, newP.importedInterfaces.Get(path, "Foreign"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	pm := map[string]string{base + "/other/ersatz": "ersatz", base + "/other/log": "otherlog", "log": "log"}
	expected := []string{"ersatz.Return", "otherlog.Level", "*log.Logger"}
	var got []string
	for _, m := range iface.Methods {
		got = append(got, m.Out[0].Type.String(pm, ""))
	}
	if !slices.Equal(got, expected) {
		t.Errorf("Expected methods to return %v but got %v", expected, got)
	}
}

// A package shadows its external test package for other packages, while
// a source file in either of them embeds the interfaces of its own.
func TestFileParser_ParsePackage_ShadowsTestPackage(t *testing.T) {
	dir := t.TempDir()
	for name, src := range map[string]string{
		"a.go":      "package a\n\ntype I interface{ A() }\n",
		"a_test.go": "package a_test\n\ntype I interface{ B() }\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	p := fileParser{srcDir: dir, packages: make(map[string]*fileParser)}
	newP, err := p.parsePackage(".")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	it := newP.importedInterfaces.Get(".", "I")
	if got := filepath.Base(newP.fileSet.Position(it.name.Pos()).Filename); got != "a.go" {
		t.Errorf("Expected I from a.go but got it from %s", got)
	}

	for pkgName, method := range map[string]string{"a": "A", "a_test": "B"} {
		fs := token.NewFileSet()
		file, err := parser.ParseFile(fs, "s.go", "package "+pkgName+"\n\ntype S interface{ I }\n", 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		src := fileParser{
			fileSet:            fs,
			imports:            make(map[string]importedPackage),
			packages:           make(map[string]*fileParser),
			importedInterfaces: newInterfaceCache(),
			auxInterfaces:      newInterfaceCache(),
			srcDir:             dir,
		}
		pkg, err := src.parseFile(".", file)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		var methods []string
		for _, m := range pkg.Interfaces[0].Methods {
			methods = append(methods, m.Name)
		}
		if !slices.Equal(methods, []string{method}) {
			t.Errorf("Expected %s.S to have methods [%s] but got %v", pkgName, method, methods)
		}
	}
}

func TestImportsOfFile(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	imports, _ := importsOfFile(file.Imports)
	checkGreeterImports(t, imports)
}

func checkGreeterImports(t *testing.T, imports map[string]importedPackage) {
	// check that imports have stdlib package "fmt"
	if fmtPackage, ok := imports["fmt"]; !ok {
		t.Errorf("Expected imports to have key \"fmt\"")
	} else {
		expectedFmtPackage := "fmt"
		if fmtPackage.Path() != expectedFmtPackage {
			t.Errorf("Expected fmt key to have value %s but got %s", expectedFmtPackage, fmtPackage.Path())
		}
	}

	// check that imports have package named "validator"
	if validatorPackage, ok := imports["validator"]; !ok {
		t.Errorf("Expected imports to have key \"fmt\"")
	} else {
		expectedValidatorPackage := "go.uber.org/mock/mockgen/internal/tests/custom_package_name/validator"
		if validatorPackage.Path() != expectedValidatorPackage {
			t.Errorf("Expected validator key to have value %s but got %s", expectedValidatorPackage, validatorPackage.Path())
		}
	}

	// check that imports have package named "client"
	if clientPackage, ok := imports["client"]; !ok {
		t.Errorf("Expected imports to have key \"client\"")
	} else {
		expectedClientPackage := "go.uber.org/mock/mockgen/internal/tests/custom_package_name/client/v1"
		if clientPackage.Path() != expectedClientPackage {
			t.Errorf("Expected client key to have value %s but got %s", expectedClientPackage, clientPackage.Path())
		}
	}

	// check that imports don't have package named "v1"
	if _, ok := imports["v1"]; ok {
		t.Errorf("Expected import not to have key \"v1\"")
	}
}

func Benchmark_parseFile(b *testing.B) {
	source := "internal/tests/performance/big_interface/big_interface.go"
	for n := 0; n < b.N; n++ {
		sourceMode(source)
	}
}

func TestParseArrayWithConstLength(t *testing.T) {
	fs := token.NewFileSet()
	srcDir := "internal/tests/const_array_length/input.go"

	file, err := parser.ParseFile(fs, srcDir, nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		auxInterfaces:      newInterfaceCache(),
		srcDir:             srcDir,
	}

	pkg, err := p.parseFile("", file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expects := []string{"[2]int", "[2]int", "[127]int", "[3]int", "[3]int", "[7]int"}
	for i, e := range expects {
		got := pkg.Interfaces[0].Methods[i].Out[0].Type.String(nil, "")
		if got != e {
			t.Fatalf("got %v; expected %v", got, e)
		}
	}
}

func TestParseFile_IncludeOnlyRequested(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		// include только один интерфейс
		includeNamesSet: map[string]struct{}{"InputMaker": {}},
	}

	pkg, err := p.parseFile("", file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkg.Interfaces) != 1 || pkg.Interfaces[0].Name != "InputMaker" {
		t.Fatalf("Expected only InputMaker, got %v", pkg.Interfaces)
	}
}

// When requested interface is missing, parser should ignore it (no error, no interfaces).
func TestParseFile_IncludeMissing_Ignored(t *testing.T) {
    fs := token.NewFileSet()
    file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }

    p := fileParser{
        fileSet:            fs,
        imports:            make(map[string]importedPackage),
        importedInterfaces: newInterfaceCache(),
        includeNamesSet:    map[string]struct{}{"DoesNotExist": {}},
    }

    pkg, err := p.parseFile("", file)
    if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(pkg.Interfaces) != 0 {
		t.Fatalf("Expected no interfaces, got %v", pkg.Interfaces)
	}
}

func TestParseFile_IncludeWithDuplicates_Dedupes(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Эмулируем «случайно указали дубликаты» как это делает sourceMode (через позиционные аргументы)
	args := []string{"InputMaker", "InputMaker"} // дубликаты
	include := make(map[string]struct{})
	for _, a := range args {
		for _, name := range strings.Split(a, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				include[name] = struct{}{}
			}
		}
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		includeNamesSet:    include,
	}

	pkg, err := p.parseFile("", file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkg.Interfaces) != 1 || pkg.Interfaces[0].Name != "InputMaker" {
		t.Fatalf("Expected only InputMaker once, got %v", pkg.Interfaces)
	}
}
