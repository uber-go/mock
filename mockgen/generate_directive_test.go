package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateDirectiveFromDestination(t *testing.T) {
	dir := t.TempDir()
	run := func(t *testing.T, dir, command string, args ...string) {
		t.Helper()
		cmd := exec.Command(command, args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", command, args, err, output)
		}
	}

	mockgen := filepath.Join(dir, "bin", "mockgen")
	if runtime.GOOS == "windows" {
		mockgen += ".exe"
	}
	run(t, "", "go", "build", "-buildvcs=false", "-o", mockgen, ".")

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "source", args: []string{"-source", "input.go", "-copyright_file=copyright notice"}},
		{name: "package", args: []string{".", "Service"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write := func(name, contents string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600); err != nil {
					t.Fatal(err)
				}
			}
			write("go.mod", "module example.com/regenerate\n\ngo 1.25.0\n")
			write("input.go", "package source\n\ntype Service interface { Work() }\n")
			write("copyright notice", "Example copyright")

			destination := filepath.Join("mocks", "nested", "mock.go")
			args := append([]string{"-write_generate_directive", "-destination", destination}, tc.args...)
			run(t, dir, mockgen, args...)

			// Change the interface so successful regeneration must update the mock.
			write("input.go", "package source\n\ntype Service interface { Regenerated() }\n")
			run(t, filepath.Join(dir, filepath.Dir(destination)), "go", "generate")

			got, err := os.ReadFile(filepath.Join(dir, destination))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), "func (m *MockService) Regenerated()") {
				t.Errorf("mock was not regenerated at %s", destination)
			}
			if tc.name == "source" && !strings.Contains(string(got), "// Example copyright") {
				t.Error("regenerated mock lost the copyright header")
			}
		})
	}
}
