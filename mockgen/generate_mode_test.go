package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseGenerateDirectiveExpandsGoGenerateVariables(t *testing.T) {
	target, ok, err := parseGenerateDirective("//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock_$GOPACKAGE Service", generateDirectiveEnv{
		GOFILE:    "service.go",
		GOLINE:    "7",
		GOPACKAGE: "service",
		DOLLAR:    "$",
	})
	if err != nil {
		t.Fatalf("parseGenerateDirective returned error: %v", err)
	}
	if !ok {
		t.Fatal("parseGenerateDirective did not recognize mockgen directive")
	}

	want := []string{"-source=service.go", "-destination=mock/service.go", "-package=mock_service", "Service"}
	if !reflect.DeepEqual(target.args, want) {
		t.Fatalf("args = %#v, want %#v", target.args, want)
	}
}

func TestParseGenerateDirectiveIgnoresNonMockgenCommands(t *testing.T) {
	_, ok, err := parseGenerateDirective("//go:generate stringer -type=Kind", generateDirectiveEnv{})
	if err != nil {
		t.Fatalf("parseGenerateDirective returned error: %v", err)
	}
	if ok {
		t.Fatal("parseGenerateDirective recognized non-mockgen directive")
	}
}

func TestParseGenerateDirectiveRecognizesGoRunWithFlags(t *testing.T) {
	target, ok, err := parseGenerateDirective("//go:generate go run -mod=mod -tags=tools go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/service_mock.go", generateDirectiveEnv{GOFILE: "service.go"})
	if err != nil {
		t.Fatalf("parseGenerateDirective returned error: %v", err)
	}
	if !ok {
		t.Fatal("parseGenerateDirective did not recognize go run mockgen directive")
	}
	want := []string{"-source=service.go", "-destination=mock/service_mock.go"}
	if !reflect.DeepEqual(target.args, want) {
		t.Fatalf("args = %#v, want %#v", target.args, want)
	}
}

func TestParseGenerateDirectiveRecognizesCommandAlias(t *testing.T) {
	aliases := generateCommandAliases{}
	if target, ok, err := parseGenerateDirectiveWithAliases("//go:generate -command mockgen-local go run -mod=mod go.uber.org/mock/mockgen", generateDirectiveEnv{}, aliases); err != nil {
		t.Fatalf("parseGenerateDirectiveWithAliases returned error: %v", err)
	} else if ok {
		t.Fatalf("-command directive returned target: %#v", target)
	}

	target, ok, err := parseGenerateDirectiveWithAliases("//go:generate mockgen-local -source=$GOFILE -destination=mock/service_mock.go", generateDirectiveEnv{GOFILE: "service.go"}, aliases)
	if err != nil {
		t.Fatalf("parseGenerateDirectiveWithAliases returned error: %v", err)
	}
	if !ok {
		t.Fatal("parseGenerateDirectiveWithAliases did not recognize alias mockgen directive")
	}
	want := []string{"-source=service.go", "-destination=mock/service_mock.go"}
	if !reflect.DeepEqual(target.args, want) {
		t.Fatalf("args = %#v, want %#v", target.args, want)
	}
}

func TestDiscoverGenerateTargetsFindsMockgenDirectives(t *testing.T) {
	dir := writeGenerateModeModule(t)

	targets, err := discoverGenerateTargets(dir, []string{"./..."})
	if err != nil {
		t.Fatalf("discoverGenerateTargets returned error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("len(targets) = %d, want 1", len(targets))
	}
	if targets[0].dir != dir {
		t.Fatalf("target dir = %q, want %q", targets[0].dir, dir)
	}
	want := []string{"-source=service.go", "-destination=mock/service_mock.go", "-package=mock_service", "Service"}
	if !reflect.DeepEqual(targets[0].args, want) {
		t.Fatalf("args = %#v, want %#v", targets[0].args, want)
	}
}

func TestDiscoverGenerateTargetsIncludesGenerateTagFiles(t *testing.T) {
	dir := writeGenerateModeModule(t)
	if err := os.WriteFile(filepath.Join(dir, "generate_only.go"), []byte(`//go:build generate

package service

//go:generate mockgen -source=$GOFILE -destination=mock/generate_only_mock.go -package=mock_service GenerateOnly

type GenerateOnly interface { OnlyGenerate() }
`), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := discoverGenerateTargets(dir, []string{"./..."})
	if err != nil {
		t.Fatalf("discoverGenerateTargets returned error: %v", err)
	}
	want := []string{"-source=generate_only.go", "-destination=mock/generate_only_mock.go", "-package=mock_service", "GenerateOnly"}
	if !hasGenerateTargetArgs(targets, want) {
		t.Fatalf("generate-tag directive not discovered: %#v", targets)
	}
}

func TestDiscoverGenerateTargetsIncludesExternalTestFiles(t *testing.T) {
	dir := writeGenerateModeModule(t)
	if err := os.WriteFile(filepath.Join(dir, "service_external_test.go"), []byte(`package service_test

//go:generate mockgen -source=$GOFILE -destination=mock/$GOPACKAGE.go -package=mock_service_test External

type External interface { DoExternal() }
`), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := discoverGenerateTargets(dir, []string{"./..."})
	if err != nil {
		t.Fatalf("discoverGenerateTargets returned error: %v", err)
	}
	want := []string{"-source=service_external_test.go", "-destination=mock/service_test.go", "-package=mock_service_test", "External"}
	if !hasGenerateTargetArgs(targets, want) {
		t.Fatalf("external test directive not discovered with GOPACKAGE=service_test: %#v", targets)
	}
}

func TestRunGenerateModeGeneratesDiscoveredMockgenDirectives(t *testing.T) {
	dir := writeGenerateModeModule(t)
	withCwd(t, dir)

	if err := runGenerateMode([]string{"./..."}); err != nil {
		t.Fatalf("runGenerateMode returned error: %v", err)
	}

	generated, err := os.ReadFile(filepath.Join(dir, "mock", "service_mock.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "type MockService") {
		t.Fatalf("generated mock does not contain MockService:\n%s", generated)
	}
}

func writeGenerateModeModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/project\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "service.go"), []byte(`package service

//go:generate mockgen -source=$GOFILE -destination=mock/service_mock.go -package=mock_service Service
//go:generate stringer -type=Kind

type Service interface { Do() }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func hasGenerateTargetArgs(targets []generateTargetDirective, want []string) bool {
	for _, target := range targets {
		if reflect.DeepEqual(target.args, want) {
			return true
		}
	}
	return false
}

func withCwd(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	})
}
