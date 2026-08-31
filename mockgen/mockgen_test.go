package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/mock/mockgen/model"
)

func TestMakeArgString(t *testing.T) {
	testCases := []struct {
		argNames  []string
		argTypes  []string
		argString string
	}{
		{
			argNames:  nil,
			argTypes:  nil,
			argString: "",
		},
		{
			argNames:  []string{"arg0"},
			argTypes:  []string{"int"},
			argString: "arg0 int",
		},
		{
			argNames:  []string{"arg0", "arg1"},
			argTypes:  []string{"int", "bool"},
			argString: "arg0 int, arg1 bool",
		},
		{
			argNames:  []string{"arg0", "arg1"},
			argTypes:  []string{"int", "int"},
			argString: "arg0, arg1 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2"},
			argTypes:  []string{"bool", "int", "int"},
			argString: "arg0 bool, arg1, arg2 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2"},
			argTypes:  []string{"int", "bool", "int"},
			argString: "arg0 int, arg1 bool, arg2 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2"},
			argTypes:  []string{"int", "int", "bool"},
			argString: "arg0, arg1 int, arg2 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2"},
			argTypes:  []string{"int", "int", "int"},
			argString: "arg0, arg1, arg2 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3"},
			argTypes:  []string{"bool", "int", "int", "int"},
			argString: "arg0 bool, arg1, arg2, arg3 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3"},
			argTypes:  []string{"int", "bool", "int", "int"},
			argString: "arg0 int, arg1 bool, arg2, arg3 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3"},
			argTypes:  []string{"int", "int", "bool", "int"},
			argString: "arg0, arg1 int, arg2 bool, arg3 int",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3"},
			argTypes:  []string{"int", "int", "int", "bool"},
			argString: "arg0, arg1, arg2 int, arg3 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3", "arg4"},
			argTypes:  []string{"bool", "int", "int", "int", "bool"},
			argString: "arg0 bool, arg1, arg2, arg3 int, arg4 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3", "arg4"},
			argTypes:  []string{"int", "bool", "int", "int", "bool"},
			argString: "arg0 int, arg1 bool, arg2, arg3 int, arg4 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3", "arg4"},
			argTypes:  []string{"int", "int", "bool", "int", "bool"},
			argString: "arg0, arg1 int, arg2 bool, arg3 int, arg4 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3", "arg4"},
			argTypes:  []string{"int", "int", "int", "bool", "bool"},
			argString: "arg0, arg1, arg2 int, arg3, arg4 bool",
		},
		{
			argNames:  []string{"arg0", "arg1", "arg2", "arg3", "arg4"},
			argTypes:  []string{"int", "int", "bool", "bool", "int"},
			argString: "arg0, arg1 int, arg2, arg3 bool, arg4 int",
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			s := makeArgString(tc.argNames, tc.argTypes)
			if s != tc.argString {
				t.Errorf("result == %q, want %q", s, tc.argString)
			}
		})
	}
}

func TestNewIdentifierAllocator(t *testing.T) {
	a := newIdentifierAllocator([]string{"taken1", "taken2"})
	if len(a) != 2 {
		t.Fatalf("expected 2 items, got %v", len(a))
	}

	_, ok := a["taken1"]
	if !ok {
		t.Errorf("allocator doesn't contain 'taken1': %#v", a)
	}

	_, ok = a["taken2"]
	if !ok {
		t.Errorf("allocator doesn't contain 'taken2': %#v", a)
	}
}

func allocatorContainsIdentifiers(a identifierAllocator, ids []string) bool {
	if len(a) != len(ids) {
		return false
	}

	for _, id := range ids {
		_, ok := a[id]
		if !ok {
			return false
		}
	}

	return true
}

func TestIdentifierAllocator_allocateIdentifier(t *testing.T) {
	a := newIdentifierAllocator([]string{"taken"})

	t2 := a.allocateIdentifier("taken_2")
	if t2 != "taken_2" {
		t.Fatalf("expected 'taken_2', got %q", t2)
	}
	expected := []string{"taken", "taken_2"}
	if !allocatorContainsIdentifiers(a, expected) {
		t.Fatalf("allocator doesn't contain the expected items - allocator: %#v, expected items: %#v", a, expected)
	}

	t3 := a.allocateIdentifier("taken")
	if t3 != "taken_3" {
		t.Fatalf("expected 'taken_3', got %q", t3)
	}
	expected = []string{"taken", "taken_2", "taken_3"}
	if !allocatorContainsIdentifiers(a, expected) {
		t.Fatalf("allocator doesn't contain the expected items - allocator: %#v, expected items: %#v", a, expected)
	}

	t4 := a.allocateIdentifier("taken")
	if t4 != "taken_4" {
		t.Fatalf("expected 'taken_4', got %q", t4)
	}
	expected = []string{"taken", "taken_2", "taken_3", "taken_4"}
	if !allocatorContainsIdentifiers(a, expected) {
		t.Fatalf("allocator doesn't contain the expected items - allocator: %#v, expected items: %#v", a, expected)
	}

	id := a.allocateIdentifier("id")
	if id != "id" {
		t.Fatalf("expected 'id', got %q", id)
	}
	expected = []string{"taken", "taken_2", "taken_3", "taken_4", "id"}
	if !allocatorContainsIdentifiers(a, expected) {
		t.Fatalf("allocator doesn't contain the expected items - allocator: %#v, expected items: %#v", a, expected)
	}
}

func TestGenerateMockInterface_Helper(t *testing.T) {
	for _, test := range []struct {
		Name       string
		Identifier string
		HelperLine string
		Methods    []*model.Method
	}{
		{Name: "mock", Identifier: "MockSomename", HelperLine: "m.ctrl.T.Helper()"},
		{Name: "recorder", Identifier: "MockSomenameMockRecorder", HelperLine: "mr.mock.ctrl.T.Helper()"},
		{
			Name:       "mock identifier conflict",
			Identifier: "MockSomename",
			HelperLine: "m_2.ctrl.T.Helper()",
			Methods: []*model.Method{
				{
					Name: "MethodA",
					In: []*model.Parameter{
						{
							Name: "m",
							Type: &model.NamedType{Type: "int"},
						},
					},
				},
			},
		},
		{
			Name:       "recorder identifier conflict",
			Identifier: "MockSomenameMockRecorder",
			HelperLine: "mr_2.mock.ctrl.T.Helper()",
			Methods: []*model.Method{
				{
					Name: "MethodA",
					In: []*model.Parameter{
						{
							Name: "mr",
							Type: &model.NamedType{Type: "int"},
						},
					},
				},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			g := generator{}

			if len(test.Methods) == 0 {
				test.Methods = []*model.Method{
					{Name: "MethodA"},
					{Name: "MethodB"},
				}
			}

			intf := &model.Interface{Name: "Somename"}
			for _, m := range test.Methods {
				intf.AddMethod(m)
			}

			if err := g.GenerateMockInterface(intf, "somepackage"); err != nil {
				t.Fatal(err)
			}

			lines := strings.Split(g.buf.String(), "\n")

			// T.Helper() should be the first line
			for _, method := range test.Methods {
				if strings.TrimSpace(lines[findMethod(t, test.Identifier, method.Name, lines)+1]) != test.HelperLine {
					t.Fatalf("method %s.%s did not declare itself a Helper method", test.Identifier, method.Name)
				}
			}
		})
	}
}

func findMethod(t *testing.T, identifier, methodName string, lines []string) int {
	t.Helper()
	r := regexp.MustCompile(fmt.Sprintf(`func\s+\(.+%s\)\s*%s`, identifier, methodName))
	for i, line := range lines {
		if r.MatchString(line) {
			return i
		}
	}

	t.Fatalf("unable to find 'func (m %s) %s'", identifier, methodName)
	panic("unreachable")
}

func TestGetArgNames(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		method   *model.Method
		expected []string
	}{
		{
			name: "NamedArg",
			method: &model.Method{
				In: []*model.Parameter{
					{
						Name: "firstArg",
						Type: &model.NamedType{Type: "int"},
					},
					{
						Name: "secondArg",
						Type: &model.NamedType{Type: "string"},
					},
				},
			},
			expected: []string{"firstArg", "secondArg"},
		},
		{
			name: "NotNamedArg",
			method: &model.Method{
				In: []*model.Parameter{
					{
						Name: "",
						Type: &model.NamedType{Type: "int"},
					},
					{
						Name: "",
						Type: &model.NamedType{Type: "string"},
					},
				},
			},
			expected: []string{"arg0", "arg1"},
		},
		{
			name: "MixedNameArg",
			method: &model.Method{
				In: []*model.Parameter{
					{
						Name: "firstArg",
						Type: &model.NamedType{Type: "int"},
					},
					{
						Name: "_",
						Type: &model.NamedType{Type: "string"},
					},
				},
			},
			expected: []string{"firstArg", "arg1"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			g := generator{}

			result := g.getArgNames(testCase.method, true)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Fatalf("expected %s, got %s", result, testCase.expected)
			}
		})
	}
}

func Test_goListImportNames(t *testing.T) {
	tests := []struct {
		name            string
		importPath      string
		wantPackageName string
		wantOK          bool
	}{
		{"golang package", "context", "context", true},
		{"third party", "golang.org/x/tools/present", "present", true},
		// Unresolvable: must be dropped, not inherit the previous entry's name.
		{"unresolvable omitted", "go.uber.org/mock/mockgen/internal/does_not_exist_xyz", "", false},
	}
	var importPaths []string
	for _, t := range tests {
		importPaths = append(importPaths, t.importPath)
	}
	packages, err := goListImportNames(importPaths)
	if err != nil {
		t.Fatalf("goListImportNames() error = %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPackageName, gotOk := packages[tt.importPath]
			if gotPackageName != tt.wantPackageName {
				t.Errorf("goListImportNames() gotPackageName = %v, wantPackageName = %v", gotPackageName, tt.wantPackageName)
			}
			if gotOk != tt.wantOK {
				t.Errorf("goListImportNames() gotOk = %v, wantOK = %v", gotOk, tt.wantOK)
			}
		})
	}
}

func Test_resolveImportNames(t *testing.T) {
	// No imports must not run `go list` (with no args it lists the cwd package).
	empty := &model.Package{Interfaces: []*model.Interface{{Name: "Empty"}}}
	names, err := resolveImportNames(empty)
	if err != nil {
		t.Fatalf("resolveImportNames(no imports) error = %v", err)
	}
	if names != nil {
		t.Errorf("resolveImportNames(no imports) = %v, want nil", names)
	}

	withImport := &model.Package{Interfaces: []*model.Interface{{
		Name: "I",
		Methods: []*model.Method{{
			Name: "M",
			In:   []*model.Parameter{{Type: &model.NamedType{Package: "time", Type: "Duration"}}},
		}},
	}}}
	names, err = resolveImportNames(withImport)
	if err != nil {
		t.Fatalf("resolveImportNames() error = %v", err)
	}
	if names["time"] != "time" {
		t.Errorf(`resolveImportNames()["time"] = %q, want "time"`, names["time"])
	}
}

// Test_Generate_usesPreResolvedPackageNames: Generate aliases imports from
// g.importNames — path basename "v2" but alias "bar" proves the map is used.
func Test_Generate_usesPreResolvedPackageNames(t *testing.T) {
	pkg := &model.Package{
		Name:    "greeter",
		PkgPath: "example.com/greeter",
		Interfaces: []*model.Interface{
			{
				Name: "Greeter",
				Methods: []*model.Method{
					{
						Name: "Greet",
						Out: []*model.Parameter{
							{Type: &model.NamedType{Package: "example.com/foo/v2", Type: "Message"}},
						},
					},
				},
			},
		},
	}

	g := &generator{
		importNames: map[string]string{
			"example.com/foo/v2": "bar",
		},
	}
	if err := g.Generate(pkg, "mock_greeter", ""); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	out := g.buf.String()
	if !strings.Contains(out, `bar "example.com/foo/v2"`) {
		t.Errorf("expected import aliased to pre-resolved name %q, got:\n%s", "bar", out)
	}
	if strings.Contains(out, `v2 "example.com/foo/v2"`) {
		t.Errorf("import fell back to path basename instead of pre-resolved name:\n%s", out)
	}
}

func TestParsePackageImport_FallbackGoPath(t *testing.T) {
	goPath := t.TempDir()
	expectedPkgPath := path.Join("example.com", "foo")
	srcDir := filepath.Join(goPath, "src", expectedPkgPath)
	err := os.MkdirAll(srcDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOPATH", goPath)
	t.Setenv("GO111MODULE", "on")
	pkgPath, err := parsePackageImport(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	if pkgPath != expectedPkgPath {
		t.Errorf("expect %s, got %s", expectedPkgPath, pkgPath)
	}
}

func TestParsePackageImport_FallbackMultiGoPath(t *testing.T) {
	// first gopath
	goPath := t.TempDir()
	goPathList := []string{goPath}
	expectedPkgPath := path.Join("example.com", "foo")
	srcDir := filepath.Join(goPath, "src", expectedPkgPath)
	err := os.MkdirAll(srcDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// second gopath
	goPath = t.TempDir()
	goPathList = append(goPathList, goPath)

	goPaths := strings.Join(goPathList, string(os.PathListSeparator))
	t.Setenv("GOPATH", goPaths)
	t.Setenv("GO111MODULE", "on")
	pkgPath, err := parsePackageImport(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	if pkgPath != expectedPkgPath {
		t.Errorf("expect %s, got %s", expectedPkgPath, pkgPath)
	}
}

func TestParseExcludeInterfaces(t *testing.T) {
	testCases := []struct {
		name     string
		arg      string
		expected map[string]struct{}
	}{
		{
			name:     "empty string",
			arg:      "",
			expected: nil,
		},
		{
			name:     "string without a comma",
			arg:      "arg1",
			expected: map[string]struct{}{"arg1": {}},
		},
		{
			name:     "two names",
			arg:      "arg1,arg2",
			expected: map[string]struct{}{"arg1": {}, "arg2": {}},
		},
		{
			name:     "two names with a comma at the end",
			arg:      "arg1,arg2,",
			expected: map[string]struct{}{"arg1": {}, "arg2": {}},
		},
		{
			name:     "two names with a comma at the beginning",
			arg:      ",arg1,arg2",
			expected: map[string]struct{}{"arg1": {}, "arg2": {}},
		},
		{
			name:     "commas only",
			arg:      ",,,,",
			expected: nil,
		},
		{
			name:     "duplicates",
			arg:      "arg1,arg2,arg1",
			expected: map[string]struct{}{"arg1": {}, "arg2": {}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseExcludeInterfaces(tt.arg)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("expected %v, actual %v", tt.expected, actual)
			}
		})
	}
}
