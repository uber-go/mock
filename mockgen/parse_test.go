package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"go.uber.org/mock/mockgen/model"
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

func TestParseFile_AnyTypeAlias(t *testing.T) {
	fs := token.NewFileSet()
	src := `package test

type Empty any
type EmptyAlias = any
type AnyMethod interface {
	any
	Foo(x any) any
	Bar(a ...any) []any
	Baz(m map[string]any) <-chan any
}
`
	file, err := parser.ParseFile(fs, "test.go", src, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		auxInterfaces:      newInterfaceCache(),
	}

	pkg, err := p.parseFile("test", file)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkg.Interfaces) != 3 {
		t.Fatalf("Expected 3 interfaces, got %d", len(pkg.Interfaces))
	}

	ifaces := make(map[string]*model.Interface)
	for _, iface := range pkg.Interfaces {
		ifaces[iface.Name] = iface
	}

	empty, ok := ifaces["Empty"]
	if !ok {
		t.Fatal("Missing interface Empty")
	}
	if len(empty.Methods) != 0 {
		t.Errorf("Expected Empty to have 0 methods, got %d", len(empty.Methods))
	}

	emptyAlias, ok := ifaces["EmptyAlias"]
	if !ok {
		t.Fatal("Missing interface EmptyAlias")
	}
	if len(emptyAlias.Methods) != 0 {
		t.Errorf("Expected EmptyAlias to have 0 methods, got %d", len(emptyAlias.Methods))
	}

	anyMethod, ok := ifaces["AnyMethod"]
	if !ok {
		t.Fatal("Missing interface AnyMethod")
	}
	if len(anyMethod.Methods) != 3 {
		t.Fatalf("Expected AnyMethod to have 3 methods, got %d", len(anyMethod.Methods))
	}

	// Verify Foo(x any) any
	foo := anyMethod.Methods[0]
	if foo.Name != "Foo" {
		t.Errorf("Expected method Foo, got %s", foo.Name)
	}
	if len(foo.In) != 1 || foo.In[0].Type != model.PredeclaredType("any") {
		t.Errorf("Expected Foo.In[0] to be any, got %v", foo.In)
	}
	if len(foo.Out) != 1 || foo.Out[0].Type != model.PredeclaredType("any") {
		t.Errorf("Expected Foo.Out[0] to be any, got %v", foo.Out)
	}

	// Verify Bar(a ...any) []any
	bar := anyMethod.Methods[1]
	if bar.Name != "Bar" {
		t.Errorf("Expected method Bar, got %s", bar.Name)
	}
	if bar.Variadic == nil || bar.Variadic.Type != model.PredeclaredType("any") {
		t.Errorf("Expected Bar.Variadic to be any, got %v", bar.Variadic)
	}
	expectedBarOut := &model.ArrayType{Len: -1, Type: model.PredeclaredType("any")}
	if len(bar.Out) != 1 || bar.Out[0].Type.String(nil, "") != expectedBarOut.String(nil, "") {
		t.Errorf("Expected Bar.Out[0] to be []any, got %v", bar.Out)
	}

	// Verify Baz(m map[string]any) <-chan any
	baz := anyMethod.Methods[2]
	if baz.Name != "Baz" {
		t.Errorf("Expected method Baz, got %s", baz.Name)
	}
	expectedBazIn := &model.MapType{Key: model.PredeclaredType("string"), Value: model.PredeclaredType("any")}
	if len(baz.In) != 1 || baz.In[0].Type.String(nil, "") != expectedBazIn.String(nil, "") {
		t.Errorf("Expected Baz.In[0] to be map[string]any, got %v", baz.In)
	}
	expectedBazOut := &model.ChanType{Dir: model.RecvDir, Type: model.PredeclaredType("any")}
	if len(baz.Out) != 1 || baz.Out[0].Type.String(nil, "") != expectedBazOut.String(nil, "") {
		t.Errorf("Expected Baz.Out[0] to be <-chan any, got %v", baz.Out)
	}
}
