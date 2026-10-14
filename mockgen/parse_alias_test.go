package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile_EmbeddedAliases(t *testing.T) {
	tests := []struct {
		name   string
		source string
		inputs []string
		output string
	}{
		{
			name: "instantiated interface",
			source: `
type Base[T any] interface { Get() T }
type Alias = Base[int]
type Target interface { Alias }
`,
			output: "int",
		},
		{
			name: "multiple type arguments",
			source: `
type Base[K, V any] interface { Get(K) V }
type Alias = Base[int, string]
type Target interface { Alias }
`,
			inputs: []string{"int"},
			output: "string",
		},
		{
			name: "forward alias chain",
			source: `
type Target interface { Chained }
type Chained = Alias
type Alias = Base[int]
type Base[T any] interface { Get() T }
`,
			output: "int",
		},
		{
			name: "parenthesized alias",
			source: `
type Base[T any] interface { Get() T }
type Alias = (Base[int])
type Target interface { Alias }
`,
			output: "int",
		},
		{
			name: "generic alias",
			source: `
type Base[T any] interface { Get() T }
type Alias[T any] = Base[[]T]
type Target interface { Alias[int] }
`,
			output: "[]int",
		},
		{
			name: "ordinary interface",
			source: `
type Base interface { Get() string }
type Alias = Base
type Target interface { Alias }
`,
			output: "string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := token.NewFileSet()
			file, err := parser.ParseFile(fs, "input.go", "package example\n"+tt.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			p := fileParser{
				fileSet:            fs,
				imports:            make(map[string]importedPackage),
				importedInterfaces: newInterfaceCache(),
				auxInterfaces:      newInterfaceCache(),
			}
			p.addAuxInterfacesFromFile("example", file)
			pkg, err := p.parseFile("example", file)
			if err != nil {
				t.Fatal(err)
			}
			target := -1
			for i, iface := range pkg.Interfaces {
				if iface.Name == "Target" {
					target = i
				}
			}
			if target < 0 {
				t.Fatal("Target interface not found")
			}
			methods := pkg.Interfaces[target].Methods
			if len(methods) != 1 || methods[0].Name != "Get" {
				t.Fatalf("expected Get method, got %v", methods)
			}
			m := methods[0]
			if len(m.In) != len(tt.inputs) {
				t.Fatalf("got %d inputs, want %d", len(m.In), len(tt.inputs))
			}
			for i, want := range tt.inputs {
				if got := m.In[i].Type.String(nil, "example"); got != want {
					t.Errorf("input %d = %s, want %s", i, got, want)
				}
			}
			if len(m.Out) != 1 {
				t.Fatalf("got %d outputs, want 1", len(m.Out))
			}
			if got := m.Out[0].Type.String(nil, "example"); got != tt.output {
				t.Errorf("output = %s, want %s", got, tt.output)
			}
		})
	}
}

func TestParseFile_EmbeddedAliasAuxFiles(t *testing.T) {
	auxFile := filepath.Join(t.TempDir(), "aliases.go")
	if err := os.WriteFile(auxFile, []byte(`package dep
type Base[T any] interface { Get() T }
type Alias = Base[int]
`), 0o600); err != nil {
		t.Fatal(err)
	}
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "input.go", `package example
import dep "example.test/dep"
type Target interface { dep.Alias }
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		auxInterfaces:      newInterfaceCache(),
	}
	if err := p.parseAuxFiles("dep=" + auxFile); err != nil {
		t.Fatal(err)
	}
	pkg, err := p.parseFile("example", file)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkg.Interfaces) != 1 || len(pkg.Interfaces[0].Methods) != 1 {
		t.Fatalf("expected one interface with one method, got %v", pkg.Interfaces)
	}
	if got := pkg.Interfaces[0].Methods[0].Out[0].Type.String(nil, "example"); got != "int" {
		t.Errorf("output = %s, want int", got)
	}
}

func TestParseFile_EmbeddedAliasDiscovery(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "input.go", `package example
type Base[T any] interface { Get() T }
type Alias = Base[int]
type Number = int
type Slice = []int
type Data[T any] struct { Value T }
type DataAlias = Data[int]
type First interface { Alias }
type Second interface { Base[string] }
type Third interface { Alias }
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		auxInterfaces:      newInterfaceCache(),
	}
	p.addAuxInterfacesFromFile("example", file)
	pkg, err := p.parseFile("example", file)
	if err != nil {
		t.Fatal(err)
	}
	// Aliases must be available for embedding without becoming extra mocks.
	if len(pkg.Interfaces) != 4 {
		t.Fatalf("got %d interfaces, want 4", len(pkg.Interfaces))
	}
	for i, want := range []string{"T", "int", "string", "int"} {
		if got := pkg.Interfaces[i].Methods[0].Out[0].Type.String(nil, "example"); got != want {
			t.Errorf("interface %s output = %s, want %s", pkg.Interfaces[i].Name, got, want)
		}
	}
}

func TestParseFile_EmbeddedAliasCycle(t *testing.T) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "input.go", `package example
type A = B
type B = A
type Target interface { A }
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	p := fileParser{
		fileSet:            fs,
		imports:            make(map[string]importedPackage),
		importedInterfaces: newInterfaceCache(),
		auxInterfaces:      newInterfaceCache(),
	}
	p.addAuxInterfacesFromFile("example", file)
	_, err = p.parseFile("example", file)
	if err == nil || !strings.Contains(err.Error(), "cyclic") {
		t.Fatalf("got error %v, want a cyclic embedding error", err)
	}
}
