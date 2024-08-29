package main

import (
	"go/parser"
	"go/token"
	"testing"
	"reflect"
	
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

func Test_filterInterfaces(t *testing.T) {
	type args struct {
		all       []*model.Interface
		requested []string
	}
	tests := []struct {
		name    string
		args    args
		want    []*model.Interface
		wantErr bool
	}{
		{
			name: "no filter",
			args: args{
				all: []*model.Interface{
					{
						Name: "Foo",
					},
					{
						Name: "Bar",
					},
				},
				requested: []string{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "filter by Foo",
			args: args{
				all: []*model.Interface{
					{
						Name: "Foo",
					},
					{
						Name: "Bar",
					},
				},
				requested: []string{"Foo"},
			},
			want: []*model.Interface{
				{
					Name: "Foo",
				},
			},
			wantErr: false,
		},
		{
			name: "filter by Foo and Bar",
			args: args{
				all: []*model.Interface{
					{
						Name: "Foo",
					},
					{
						Name: "Bar",
					},
				},
				requested: []string{"Foo", "Bar"},
			},
			want: []*model.Interface{
				{
					Name: "Foo",
				},
				{
					Name: "Bar",
				},
			},
			wantErr: false,
		},
		{
			name: "incorrect filter by Foo and Baz",
			args: args{
				all: []*model.Interface{
					{
						Name: "Foo",
					},
					{
						Name: "Bar",
					},
				},
				requested: []string{"Foo", "Baz"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filterInterfaces(tt.args.all, tt.args.requested)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterInterfaces() got = %v, want %v", got, tt.want)
			}
		})
	}
}