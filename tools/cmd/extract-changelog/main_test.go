package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _changelog = `
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- Upcoming feature

## [1.0.0] - 2021-08-18
Initial stable release.

[1.0.0]: http://example.com/1.0.0

## 0.3.0 - 2020-09-01
### Removed
- deprecated functionality

### Fixed
- bug

## [0.2.0] - 2020-08-19
### Added
- Fancy new feature.

[0.2.0]: http://example.com/0.2.0

## 0.1.0 - 2020-08-18

Initial release.
`

func TestMain(t *testing.T) {
	t.Parallel()

	changelog := filepath.Join(t.TempDir(), "CHANGELOG.md")
	require.NoError(t,
		os.WriteFile(changelog, []byte(_changelog), 0o644))

	tests := []struct {
		desc string

		version string
		want    string // expected changelog
		wantErr string // expected error, if any
	}{
		{
			desc:    "not found",
			version: "0.1.1",
			wantErr: `changelog for "0.1.1" not found`,
		},
		{
			desc:    "missing version",
			wantErr: "please provide a version",
		},
		{
			desc:    "non-standard body",
			version: "1.0.0",
			want: joinLines(
				"## [1.0.0] - 2021-08-18",
				"Initial stable release.",
				"",
				"[1.0.0]: http://example.com/1.0.0",
			),
		},
		{
			desc:    "unlinked",
			version: "0.3.0",
			want: joinLines(
				"## 0.3.0 - 2020-09-01",
				"### Removed",
				"- deprecated functionality",
				"",
				"### Fixed",
				"- bug",
			),
		},
		{
			desc:    "end of file",
			version: "0.1.0",
			want: joinLines(
				"## 0.1.0 - 2020-08-18",
				"",
				"Initial release.",
			),
		},
		{
			desc:    "linked",
			version: "0.2.0",
			want: joinLines(
				"## [0.2.0] - 2020-08-19",
				"### Added",
				"- Fancy new feature.",
				"",
				"[0.2.0]: http://example.com/0.2.0",
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			defer func() {
				assert.Empty(t, stderr.String(), "stderr should be empty")
			}()

			err := (&mainCmd{
				Stdout: &stdout,
				Stderr: &stderr,
			}).Run([]string{"-i", changelog, tt.version})

			if len(tt.wantErr) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, stdout.String())
		})
	}
}

// Join a bunch of lines with a trailing newline.
func joinLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
