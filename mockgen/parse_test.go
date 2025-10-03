package main

import (
	"go/parser"
	"go/token"
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
		importedInterfaces: newInterfaceCache(),
	}

	newP, err := p.parsePackage("go.uber.org/mock/mockgen/internal/tests/custom_package_name/greeter")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	checkGreeterImports(t, newP.imports)
}

func TestImportsOfFile(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "internal/tests/custom_package_name/greeter/greeter.go", nil, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	imports, _ := importsOfFile(file)
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

func TestParseFile_IncludeMissing_ReturnsError(t *testing.T) {
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

	_, err = p.parseFile("", file)
	if err == nil || !strings.Contains(err.Error(), "requested interfaces not found") {
		t.Fatalf("Expected missing interface error, got %v", err)
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
