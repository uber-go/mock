// extract-changelog extracts the release notes for a specific version from a
// file matching the format prescribed by https://keepachangelog.com/en/1.0.0/.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	cmd := mainCmd{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := cmd.Run(os.Args[1:]); err != nil && err != flag.ErrHelp {
		fmt.Fprintln(cmd.Stderr, err)
		os.Exit(1)
	}
}

type mainCmd struct {
	Stdout io.Writer
	Stderr io.Writer
}

const _usage = `USAGE

	%v [OPTIONS] VERSION

Retrieves the release notes for VERSION from a CHANGELOG.md file and prints
them to stdout.

EXAMPLES

  extract-changelog -i CHANGELOG.md v1.2.3
  extract-changelog 0.2.5

OPTIONS
`

func (cmd *mainCmd) Run(args []string) error {
	flag := flag.NewFlagSet("extract-changelog", flag.ContinueOnError)
	flag.SetOutput(cmd.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(flag.Output(), _usage, flag.Name())
		flag.PrintDefaults()
	}

	file := flag.String("i", "CHANGELOG.md", "input file")

	if err := flag.Parse(args); err != nil {
		return err
	}

	var version string
	if args := flag.Args(); len(args) > 0 {
		version = args[0]
	}
	version = strings.TrimPrefix(version, "v")

	if len(version) == 0 {
		return errors.New("please provide a version")
	}

	f, err := os.Open(*file)
	if err != nil {
		return fmt.Errorf("open changelog: %v", err)
	}
	defer f.Close()

	s, err := extract(f, version)
	if err != nil {
		return err
	}
	_, err = io.WriteString(cmd.Stdout, s)
	return err
}

func extract(r io.Reader, version string) (string, error) {
	type _state int

	const (
		initial _state = iota
		foundHeader
	)

	var (
		state   _state
		buff    bytes.Buffer
		scanner = bufio.NewScanner(r)
	)

scan:
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case initial:
			// Version headers take one of the following forms:
			//
			//   ## 0.1.3 - 2021-08-18
			//   ## [0.1.3] - 2021-08-18
			switch {
			case strings.HasPrefix(line, "## "+version+" "),
				strings.HasPrefix(line, "## ["+version+"]"):
				fmt.Fprintln(&buff, line)
				state = foundHeader
			}

		case foundHeader:
			// Found a new version header. Stop extracting.
			if strings.HasPrefix(line, "## ") {
				break scan
			}
			fmt.Fprintln(&buff, line)

		default:
			// unreachable but guard against it.
			return "", fmt.Errorf("unexpected state %v at %q", state, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if state < foundHeader {
		return "", fmt.Errorf("changelog for %q not found", version)
	}

	out := buff.String()
	out = strings.TrimSpace(out) + "\n" // always end with a single newline
	return out, nil
}
