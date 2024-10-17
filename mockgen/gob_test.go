package main

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGobMode(t *testing.T) {

	// Encode a package to a temporary gob.
	parser := packageModeParser{}
	want, err := parser.parsePackage(
		"go.uber.org/mock/mockgen/internal/tests/package_mode" /* package name */,
		[]string{ "Human", "Earth" } /* ifaces */,
	)
	path := filepath.Join(t.TempDir(), "model.gob")
	outfile, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, gob.NewEncoder(outfile).Encode(want))
	outfile.Close()

	// Ensure gobMode loads it correctly.
	got, err := gobMode(path)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
