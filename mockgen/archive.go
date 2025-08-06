package main

import (
	"fmt"
	"go/token"
	"go/types"
	"os"

	"go.uber.org/mock/mockgen/model"

	"golang.org/x/tools/go/gcexportdata"
)

func parseExportFile(importPath string, symbols []string, archive string) (*model.Package, error) {
	f, err := os.Open(archive)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r, err := gcexportdata.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("read export data %q: %v", archive, err)
	}

	fset := token.NewFileSet()
	imports := make(map[string]*types.Package)
	tp, err := gcexportdata.Read(r, fset, imports, importPath)
	if err != nil {
		return nil, err
	}

	interfaces, err := extractInterfacesFromPackageTypes(tp, symbols)
	if err != nil {
		return nil, err
	}

	pkg := &model.Package{
		Name:       tp.Name(),
		PkgPath:    tp.Path(),
		Interfaces: interfaces,
	}
	return pkg, nil
}
