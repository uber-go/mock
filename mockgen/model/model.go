// Copyright 2012 Google Inc.
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

// Package model contains the data model necessary for generating mock implementations.
package model

import (
	"encoding/gob"
	"fmt"
	"io"
	"strings"
)

// pkgPath is the importable path for package model
const pkgPath = "go.uber.org/mock/mockgen/model"

// Package is a Go package. It may be a subset.
type Package struct {
	Name       string
	PkgPath    string
	Interfaces []*Interface
	DotImports []string
}

// Print writes the package name and its exported interfaces.
func (pkg *Package) Print(w io.Writer) {
	_, _ = fmt.Fprintf(w, "package %s\n", pkg.Name)
	for _, intf := range pkg.Interfaces {
		intf.Print(w)
	}
}

// Imports returns the imports needed by the Package as a set of import paths.
func (pkg *Package) Imports() map[string]bool {
	im := make(map[string]bool)
	for _, intf := range pkg.Interfaces {
		intf.addImports(im)
		for _, tp := range intf.TypeParams {
			tp.Type.addImports(im)
		}
	}
	return im
}

// Interface is a Go interface.
type Interface struct {
	Name       string
	Methods    []*Method
	TypeParams []*Parameter
}

// Print writes the interface name and its methods.
func (intf *Interface) Print(w io.Writer) {
	_, _ = fmt.Fprintf(w, "interface %s\n", intf.Name)
	for _, m := range intf.Methods {
		m.Print(w)
	}
}

func (intf *Interface) addImports(im map[string]bool) {
	for _, m := range intf.Methods {
		m.addImports(im)
	}
}

// AddMethod adds a new method, de-duplicating by method name.
func (intf *Interface) AddMethod(m *Method) {
	for _, me := range intf.Methods {
		if me.Name == m.Name {
			return
		}
	}
	intf.Methods = append(intf.Methods, m)
}

// Method is a single method of an interface.
type Method struct {
	Name     string
	In, Out  []*Parameter
	Variadic *Parameter // may be nil
}

// Print writes the method name and its signature.
func (m *Method) Print(w io.Writer) {
	_, _ = fmt.Fprintf(w, "  - method %s\n", m.Name)
	if len(m.In) > 0 {
		_, _ = fmt.Fprintf(w, "    in:\n")
		for _, p := range m.In {
			p.Print(w)
		}
	}
	if m.Variadic != nil {
		_, _ = fmt.Fprintf(w, "    ...:\n")
		m.Variadic.Print(w)
	}
	if len(m.Out) > 0 {
		_, _ = fmt.Fprintf(w, "    out:\n")
		for _, p := range m.Out {
			p.Print(w)
		}
	}
}

func (m *Method) addImports(im map[string]bool) {
	for _, p := range m.In {
		p.Type.addImports(im)
	}
	if m.Variadic != nil {
		m.Variadic.Type.addImports(im)
	}
	for _, p := range m.Out {
		p.Type.addImports(im)
	}
}

// Parameter is an argument or return parameter of a method.
type Parameter struct {
	Name string // may be empty
	Type Type
}

// Print writes a method parameter.
func (p *Parameter) Print(w io.Writer) {
	n := p.Name
	if n == "" {
		n = `""`
	}
	_, _ = fmt.Fprintf(w, "    - %v: %v\n", n, p.Type.String(nil, ""))
}

// Type is a Go type.
type Type interface {
	String(pm map[string]string, pkgOverride string) string
	addImports(im map[string]bool)
}

func init() {
	// Call gob.RegisterName with pkgPath as prefix to avoid conflicting with
	// github.com/golang/mock/mockgen/model 's registration.
	gob.RegisterName(pkgPath+".ArrayType", &ArrayType{})
	gob.RegisterName(pkgPath+".ChanType", &ChanType{})
	gob.RegisterName(pkgPath+".FuncType", &FuncType{})
	gob.RegisterName(pkgPath+".MapType", &MapType{})
	gob.RegisterName(pkgPath+".NamedType", &NamedType{})
	gob.RegisterName(pkgPath+".PointerType", &PointerType{})

	// Call gob.RegisterName to make sure it has the consistent name registered
	// for both gob decoder and encoder.
	//
	// For a non-pointer type, gob.Register will try to get package full path by
	// calling rt.PkgPath() for a name to register. If your project has vendor
	// directory, it is possible that PkgPath will get a path like this:
	//     ../../../vendor/go.uber.org/mock/mockgen/model
	gob.RegisterName(pkgPath+".PredeclaredType", PredeclaredType(""))
}

// ArrayType is an array or slice type.
type ArrayType struct {
	Len  int // -1 for slices, >= 0 for arrays
	Type Type
}

func (at *ArrayType) String(pm map[string]string, pkgOverride string) string {
	s := "[]"
	if at.Len > -1 {
		s = fmt.Sprintf("[%d]", at.Len)
	}
	return s + at.Type.String(pm, pkgOverride)
}

func (at *ArrayType) addImports(im map[string]bool) { at.Type.addImports(im) }

// ChanType is a channel type.
type ChanType struct {
	Dir  ChanDir // 0, 1 or 2
	Type Type
}

func (ct *ChanType) String(pm map[string]string, pkgOverride string) string {
	s := ct.Type.String(pm, pkgOverride)
	if ct.Dir == RecvDir {
		return "<-chan " + s
	}
	if ct.Dir == SendDir {
		return "chan<- " + s
	}
	return "chan " + s
}

func (ct *ChanType) addImports(im map[string]bool) { ct.Type.addImports(im) }

// ChanDir is a channel direction.
type ChanDir int

// Constants for channel directions.
const (
	RecvDir ChanDir = 1
	SendDir ChanDir = 2
)

// FuncType is a function type.
type FuncType struct {
	In, Out  []*Parameter
	Variadic *Parameter // may be nil
}

func (ft *FuncType) String(pm map[string]string, pkgOverride string) string {
	args := make([]string, len(ft.In))
	for i, p := range ft.In {
		args[i] = p.Type.String(pm, pkgOverride)
	}
	if ft.Variadic != nil {
		args = append(args, "..."+ft.Variadic.Type.String(pm, pkgOverride))
	}
	rets := make([]string, len(ft.Out))
	for i, p := range ft.Out {
		rets[i] = p.Type.String(pm, pkgOverride)
	}
	retString := strings.Join(rets, ", ")
	if nOut := len(ft.Out); nOut == 1 {
		retString = " " + retString
	} else if nOut > 1 {
		retString = " (" + retString + ")"
	}
	return "func(" + strings.Join(args, ", ") + ")" + retString
}

func (ft *FuncType) addImports(im map[string]bool) {
	for _, p := range ft.In {
		p.Type.addImports(im)
	}
	if ft.Variadic != nil {
		ft.Variadic.Type.addImports(im)
	}
	for _, p := range ft.Out {
		p.Type.addImports(im)
	}
}

// MapType is a map type.
type MapType struct {
	Key, Value Type
}

func (mt *MapType) String(pm map[string]string, pkgOverride string) string {
	return "map[" + mt.Key.String(pm, pkgOverride) + "]" + mt.Value.String(pm, pkgOverride)
}

func (mt *MapType) addImports(im map[string]bool) {
	mt.Key.addImports(im)
	mt.Value.addImports(im)
}

// NamedType is an exported type in a package.
type NamedType struct {
	Package    string // may be empty
	Type       string
	TypeParams *TypeParametersType
}

func (nt *NamedType) String(pm map[string]string, pkgOverride string) string {
	if pkgOverride == nt.Package {
		return nt.Type + nt.TypeParams.String(pm, pkgOverride)
	}
	prefix := pm[nt.Package]
	if prefix != "" {
		return prefix + "." + nt.Type + nt.TypeParams.String(pm, pkgOverride)
	}

	return nt.Type + nt.TypeParams.String(pm, pkgOverride)
}

func (nt *NamedType) addImports(im map[string]bool) {
	if nt.Package != "" {
		im[nt.Package] = true
	}
	nt.TypeParams.addImports(im)
}

// PointerType is a pointer to another type.
type PointerType struct {
	Type Type
}

func (pt *PointerType) String(pm map[string]string, pkgOverride string) string {
	return "*" + pt.Type.String(pm, pkgOverride)
}
func (pt *PointerType) addImports(im map[string]bool) { pt.Type.addImports(im) }

// PredeclaredType is a predeclared type such as "int".
type PredeclaredType string

func (pt PredeclaredType) String(map[string]string, string) string { return string(pt) }
func (pt PredeclaredType) addImports(map[string]bool)              {}

// TypeParametersType contains type parameters for a NamedType.
type TypeParametersType struct {
	TypeParameters []Type
}

func (tp *TypeParametersType) String(pm map[string]string, pkgOverride string) string {
	if tp == nil || len(tp.TypeParameters) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range tp.TypeParameters {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(v.String(pm, pkgOverride))
	}
	sb.WriteString("]")
	return sb.String()
}

func (tp *TypeParametersType) addImports(im map[string]bool) {
	if tp == nil {
		return
	}
	for _, v := range tp.TypeParameters {
		v.addImports(im)
	}
}

// ErrorInterface represent built-in error interface.
var ErrorInterface = Interface{
	Name: "error",
	Methods: []*Method{
		{
			Name: "Error",
			Out: []*Parameter{
				{
					Name: "",
					Type: PredeclaredType("string"),
				},
			},
		},
	},
}
