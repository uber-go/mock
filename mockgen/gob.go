package main

import (
	"encoding/gob"
	"os"

	"go.uber.org/mock/mockgen/model"
)

func gobMode(path string) (*model.Package, map[string]string, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer in.Close()
	var pkg model.Package
	if err := gob.NewDecoder(in).Decode(&pkg); err != nil {
		return nil, nil, err
	}
	importNames, err := resolveImportNames(&pkg)
	if err != nil {
		return nil, nil, err
	}
	return &pkg, importNames, nil
}
