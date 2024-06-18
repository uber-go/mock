// Copyright 2010 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// MockGen generates mock implementations of Go interfaces.
package main

// TODO: This does not support recursive embedded interfaces.
// TODO: This does not support embedding package-local interfaces in a separate file.

import (
	"flag"
	"io"
	"log"
	"os"

	"go.uber.org/mock/mockgen/api"
)

const (
	gomockImportPath = "go.uber.org/mock/gomock"
)

var (
	version = ""
	commit  = "none"
	date    = "unknown"
)

var (
	source                 = flag.String("source", "", "(source mode) Input Go source file; enables source mode.")
	destination            = flag.String("destination", "", "Output file; defaults to stdout.")
	mockNames              = flag.String("mock_names", "", "Comma-separated interfaceName=mockName pairs of explicit mock names to use. Mock names default to 'Mock'+ interfaceName suffix.")
	packageOut             = flag.String("package", "", "Package of the generated code; defaults to the package of the input with a 'mock_' prefix.")
	selfPackage            = flag.String("self_package", "", "The full package import path for the generated code. The purpose of this flag is to prevent import cycles in the generated code by trying to include its own package. This can happen if the mock's package is set to one of its inputs (usually the main one) and the output is stdio so mockgen cannot detect the final output package. Setting this flag will then tell mockgen which import to exclude.")
	writeCmdComment        = flag.Bool("write_command_comment", true, "Writes the command used as a comment if true.")
	writePkgComment        = flag.Bool("write_package_comment", true, "Writes package documentation comment (godoc) if true.")
	writeSourceComment     = flag.Bool("write_source_comment", true, "Writes original file (source mode) or interface names (reflect mode) comment if true.")
	writeGenerateDirective = flag.Bool("write_generate_directive", false, "Add //go:generate directive to regenerate the mock")
	copyrightFile          = flag.String("copyright_file", "", "Copyright file used to add copyright header")
	typed                  = flag.Bool("typed", false, "Generate Type-safe 'Return', 'Do', 'DoAndReturn' function")
	imports                = flag.String("imports", "", "(source mode) Comma-separated name=path pairs of explicit imports to use.")
	auxFiles               = flag.String("aux_files", "", "(source mode) Comma-separated pkg=path pairs of auxiliary Go source files.")
	excludeInterfaces      = flag.String("exclude_interfaces", "", "Comma-separated names of interfaces to be excluded")

	debugParser = flag.Bool("debug_parser", false, "Print out parser results only.")
	showVersion = flag.Bool("version", false, "Print version.")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *source == "" {
		if flag.NArg() != 2 {
			usage()
			log.Fatal("Expected exactly two arguments")
		}
	}

	api.Generate(api.Config{
		Source:                 *source,
		Destination:            *destination,
		MockNames:              *mockNames,
		PackageOut:             *packageOut,
		SelfPackage:            *selfPackage,
		WriteCmdComment:        *writeCmdComment,
		WritePkgComment:        *writePkgComment,
		WriteSourceComment:     *writeSourceComment,
		WriteGenerateDirective: *writeGenerateDirective,
		CopyrightFile:          *copyrightFile,
		Typed:                  *typed,
		Imports:                *imports,
		AuxFiles:               *auxFiles,
		ExcludeInterfaces:      *excludeInterfaces,
		DebugParser:            *debugParser,
		ShowVersion:            *showVersion,
	})
}

func usage() {
	_, _ = io.WriteString(os.Stderr, usageText)
	flag.PrintDefaults()
}

const usageText = `mockgen has two modes of operation: source and reflect.

Source mode generates mock interfaces from a source file.
It is enabled by using the -source flag. Other flags that
may be useful in this mode are -imports and -aux_files.
Example:
	mockgen -source=foo.go [other options]

Reflect mode generates mock interfaces by building a program
that uses reflection to understand interfaces. It is enabled
by passing two non-flag arguments: an import path, and a
comma-separated list of symbols.
Example:
	mockgen database/sql/driver Conn,Driver

`
