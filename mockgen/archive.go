package main

import (
	"fmt"
	"go/token"
	"go/types"
	"os"

	"go.uber.org/mock/mockgen/model"

	"golang.org/x/tools/go/gcexportdata"
)

// parseExportFile builds the package model from an archive's export data and
// returns each import's package name, so callers need no `go list`.
func parseExportFile(importPath string, symbols []string, archive string) (*model.Package, map[string]string, error) {
	f, err := os.Open(archive)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r, err := gcexportdata.NewReader(f)
	if err != nil {
		return nil, nil, fmt.Errorf("read export data %q: %v", archive, err)
	}

	fset := token.NewFileSet()
	imports := make(map[string]*types.Package)
	tp, err := gcexportdata.Read(r, fset, imports, importPath)
	if err != nil {
		return nil, nil, err
	}

	interfaces, err := extractInterfacesFromPackageTypes(tp, symbols)
	if err != nil {
		return nil, nil, err
	}

	// Package names come from the export data; they can differ from the path
	// basename (e.g. "gopkg.in/yaml.v3" is package "yaml").
	importNames := make(map[string]string, len(imports)+1)
	importNames[tp.Path()] = tp.Name()
	for path, p := range imports {
		if p != nil && p.Name() != "" {
			importNames[path] = p.Name()
		}
	}

	pkg := &model.Package{
		Name:       tp.Name(),
		PkgPath:    tp.Path(),
		Interfaces: interfaces,
	}
	return pkg, importNames, nil
}
