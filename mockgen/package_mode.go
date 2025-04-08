package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"go.uber.org/mock/mockgen/model"
	"golang.org/x/tools/go/packages"
)

var (
	buildFlags = flag.String("build_flags", "", "(package mode) Additional flags for go build.")
)

type packageModeParser struct {
	// Mapping from underlying types to aliases used within the package source.
	//
	// We prefer to use aliases used in the source rather than underlying type names
	// as those may be unexported or internal.
	// TODO(joaks): Once mock is Go1.23+ only, we can remove this
	// as the casing for types.Alias will automatically handle this
	// in all cases.
	aliasReplacements map[types.Type]aliasReplacement
}

type aliasReplacement struct {
	name string
	pkg  string
}

func (p *packageModeParser) parsePackage(packageName string, ifaces []string) (*model.Package, error) {
	parsed, err := p.parsePackages([]parseTarget{{name: packageName, ifaces: ifaces}})
	if err != nil {
		return nil, err
	}
	return parsed[0], nil
}

type parseTarget struct {
	name   string
	ifaces []string
}

func (p *packageModeParser) parsePackages(targets []parseTarget) ([]*model.Package, error) {
	packageNames := make([]string, len(targets))
	for i := range targets {
		packageNames[i] = targets[i].name
	}

	pkgs, err := loadPackages(packageNames)
	if err != nil {
		return nil, fmt.Errorf("load package: %w", err)
	}

	p.buildAliasReplacements(pkgs)

	pkgByPath := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgByPath[pkg.PkgPath] = pkg
	}

	parsed := make([]*model.Package, len(targets))
	for i := range targets {
		pkg, ok := pkgByPath[targets[i].name]
		if !ok {
			return nil, fmt.Errorf("package not found: %s", targets[i].name)
		}
		interfaces, err := p.extractInterfacesFromPackage(pkg, targets[i].ifaces)
		if err != nil {
			return nil, fmt.Errorf("extract interfaces from package: %w", err)
		}

		parsed[i] = &model.Package{
			Name:       pkg.Types.Name(),
			PkgPath:    pkg.PkgPath,
			Interfaces: interfaces,
		}
	}
	return parsed, nil
}

// buildAliasReplacements finds and records any references to aliases
// within the given package's source.
// These aliases will be preferred when parsing types
// over the underlying name counterparts, as those may be unexported / internal.
//
// If a type has more than one alias within the source package,
// the latest one to be inspected will be the one used for mapping.
// This is fine, since all aliases and their underlying types are interchangeable
// from a type-checking standpoint.
func (p *packageModeParser) buildAliasReplacements(pkgs []*packages.Package) {
	p.aliasReplacements = make(map[types.Type]aliasReplacement)

	// checkIdent checks if the given identifier exists
	// in the given package as an alias, and adds it to
	// the alias replacements map if so.
	checkIdent := func(pkg *types.Package, ident string) bool {
		scope := pkg.Scope()
		if scope == nil {
			return true
		}
		obj := scope.Lookup(ident)
		if obj == nil {
			return true
		}
		objTypeName, ok := obj.(*types.TypeName)
		if !ok {
			return true
		}
		if !objTypeName.IsAlias() {
			return true
		}
		typ := objTypeName.Type()
		if typ == nil {
			return true
		}
		p.aliasReplacements[typ] = aliasReplacement{
			name: objTypeName.Name(),
			pkg:  pkg.Path(),
		}
		return false
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			fileScope, ok := pkg.TypesInfo.Scopes[f]
			if !ok {
				continue
			}
			ast.Inspect(f, func(node ast.Node) bool {
				// Simple identifiers: check if it is an alias
				// from the source package.
				if ident, ok := node.(*ast.Ident); ok {
					return checkIdent(pkg.Types, ident.String())
				}

				// Selector expressions: check if it is an alias
				// from the package represented by the qualifier.
				selExpr, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				x, sel := selExpr.X, selExpr.Sel
				xident, ok := x.(*ast.Ident)
				if !ok {
					return true
				}

				xObj := fileScope.Lookup(xident.String())
				pkgName, ok := xObj.(*types.PkgName)
				if !ok {
					return true
				}

				xPkg := pkgName.Imported()
				if xPkg == nil {
					return true
				}
				return checkIdent(xPkg, sel.String())
			})
		}
	}
}

func loadPackages(packageNames []string) ([]*packages.Package, error) {
	var buildFlagsSet []string
	if *buildFlags != "" {
		buildFlagsSet = strings.Split(*buildFlags, " ")
	}

	cfg := &packages.Config{
		Mode:       packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedEmbedFiles | packages.LoadSyntax,
		BuildFlags: buildFlagsSet,
	}
	pkgs, err := packages.Load(cfg, packageNames...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var errs []error
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			errs = append(errs, err)
		}
	}
	return pkgs, errors.Join(errs...)
}

func (p *packageModeParser) extractInterfacesFromPackage(pkg *packages.Package, ifaces []string) ([]*model.Interface, error) {
	interfaces := make([]*model.Interface, len(ifaces))
	for i, iface := range ifaces {
		obj := pkg.Types.Scope().Lookup(iface)
		if obj == nil {
			return nil, fmt.Errorf("interface %s does not exist", iface)
		}

		modelIface, err := p.parseInterface(obj)
		if err != nil {
			return nil, newParseTypeError("parse interface", obj.Name(), err)
		}

		interfaces[i] = modelIface
	}

	return interfaces, nil
}

func (p *packageModeParser) parseInterface(obj types.Object) (*model.Interface, error) {
	named, ok := types.Unalias(obj.Type()).(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %T", obj.Name(), obj.Type().Underlying())
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %T", obj.Name(), obj.Type().Underlying())
	}

	if p.isConstraint(iface) {
		return nil, fmt.Errorf("interface %s is a constraint", obj.Name())
	}

	methods := make([]*model.Method, iface.NumMethods())
	for i := range iface.NumMethods() {
		method := iface.Method(i)
		typedMethod, ok := method.Type().(*types.Signature)
		if !ok {
			return nil, fmt.Errorf("method %s is not a signature", method.Name())
		}

		modelFunc, err := p.parseFunc(typedMethod)
		if err != nil {
			return nil, newParseTypeError("parse method", typedMethod.String(), err)
		}

		methods[i] = &model.Method{
			Name:     method.Name(),
			In:       modelFunc.In,
			Out:      modelFunc.Out,
			Variadic: modelFunc.Variadic,
		}
	}

	if named.TypeParams() == nil {
		return &model.Interface{Name: obj.Name(), Methods: methods}, nil
	}

	typeParams := make([]*model.Parameter, named.TypeParams().Len())
	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		typeParam, err := p.parseConstraint(param)
		if err != nil {
			return nil, newParseTypeError("parse type parameter", param.String(), err)
		}

		typeParams[i] = &model.Parameter{Name: param.Obj().Name(), Type: typeParam}
	}

	return &model.Interface{Name: obj.Name(), Methods: methods, TypeParams: typeParams}, nil
}

func (p *packageModeParser) isConstraint(t *types.Interface) bool {
	for i := range t.NumEmbeddeds() {
		embed := t.EmbeddedType(i)
		if _, ok := embed.Underlying().(*types.Interface); !ok {
			return true
		}
	}

	return false
}

func (p *packageModeParser) parseType(t types.Type) (model.Type, error) {
	switch t := t.(type) {
	case *types.Array:
		elementType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse array type", t.Elem().String(), err)
		}
		return &model.ArrayType{Len: int(t.Len()), Type: elementType}, nil
	case *types.Slice:
		elementType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse slice type", t.Elem().String(), err)
		}

		return &model.ArrayType{Len: -1, Type: elementType}, nil
	case *types.Chan:
		var dir model.ChanDir
		switch t.Dir() {
		case types.RecvOnly:
			dir = model.RecvDir
		case types.SendOnly:
			dir = model.SendDir
		}

		chanType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse chan type", t.Elem().String(), err)
		}

		return &model.ChanType{Dir: dir, Type: chanType}, nil
	case *types.Signature:
		sig, err := p.parseFunc(t)
		if err != nil {
			return nil, newParseTypeError("parse signature", t.String(), err)
		}

		return sig, nil
	case *types.Named, *types.Alias:
		object := t.(interface{ Obj() *types.TypeName })
		name := object.Obj().Name()
		var pkg string
		if object.Obj().Pkg() != nil {
			pkg = object.Obj().Pkg().Path()
		}

		// If there was an alias to this type used somewhere in the source,
		// use that alias instead of the underlying type,
		// since the underlying type might be unexported.
		if alias, ok := p.aliasReplacements[t]; ok {
			name = alias.name
			pkg = alias.pkg
		}

		// TypeArgs method not available for aliases in go1.22
		genericType, ok := t.(interface{ TypeArgs() *types.TypeList })
		if !ok || genericType.TypeArgs() == nil {
			return &model.NamedType{
				Package: pkg,
				Type:    name,
			}, nil
		}

		typeParams := &model.TypeParametersType{TypeParameters: make([]model.Type, genericType.TypeArgs().Len())}
		for i := range genericType.TypeArgs().Len() {
			typeParam := genericType.TypeArgs().At(i)
			typedParam, err := p.parseType(typeParam)
			if err != nil {
				return nil, newParseTypeError("parse type parameter", typeParam.String(), err)
			}

			typeParams.TypeParameters[i] = typedParam
		}

		return &model.NamedType{
			Package:    pkg,
			Type:       name,
			TypeParams: typeParams,
		}, nil
	case *types.Interface:
		if t.Empty() {
			return model.PredeclaredType("any"), nil
		}

		return nil, fmt.Errorf("cannot handle non-empty unnamed interfaces")
	case *types.Map:
		key, err := p.parseType(t.Key())
		if err != nil {
			return nil, newParseTypeError("parse map key", t.Key().String(), err)
		}
		value, err := p.parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse map value", t.Elem().String(), err)
		}

		return &model.MapType{Key: key, Value: value}, nil
	case *types.Pointer:
		valueType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse pointer type", t.Elem().String(), err)
		}

		return &model.PointerType{Type: valueType}, nil
	case *types.Struct:
		if t.NumFields() > 0 {
			return nil, fmt.Errorf("cannot handle non-empty unnamed structs")
		}

		return model.PredeclaredType("struct{}"), nil
	case *types.Basic:
		return model.PredeclaredType(t.Name()), nil
	case *types.Tuple:
		panic("tuple field") // TODO
	case *types.TypeParam:
		return &model.NamedType{Type: t.Obj().Name()}, nil
	default:
		panic("unknown type") // TODO
	}
}

func (p *packageModeParser) parseFunc(sig *types.Signature) (*model.FuncType, error) {
	var variadic *model.Parameter
	params := make([]*model.Parameter, 0, sig.Params().Len())
	for i := range sig.Params().Len() {
		param := sig.Params().At(i)

		isVariadicParam := i == sig.Params().Len()-1 && sig.Variadic()
		parseType := param.Type()
		if isVariadicParam {
			sliceType, ok := param.Type().(*types.Slice)
			if !ok {
				return nil, newParseTypeError("variadic parameter is not a slice", param.String(), nil)
			}

			parseType = sliceType.Elem()
		}

		paramType, err := p.parseType(parseType)
		if err != nil {
			return nil, newParseTypeError("parse parameter type", parseType.String(), err)
		}

		modelParameter := &model.Parameter{Type: paramType, Name: param.Name()}

		if isVariadicParam {
			variadic = modelParameter
		} else {
			params = append(params, modelParameter)
		}
	}

	if len(params) == 0 {
		params = nil
	}

	results := make([]*model.Parameter, sig.Results().Len())
	for i := range sig.Results().Len() {
		result := sig.Results().At(i)

		resultType, err := p.parseType(result.Type())
		if err != nil {
			return nil, newParseTypeError("parse result type", result.Type().String(), err)
		}

		results[i] = &model.Parameter{Type: resultType, Name: result.Name()}
	}

	if len(results) == 0 {
		results = nil
	}

	return &model.FuncType{
		In:       params,
		Out:      results,
		Variadic: variadic,
	}, nil
}

func (p *packageModeParser) parseConstraint(t *types.TypeParam) (model.Type, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type param")
	}

	typeParam, err := p.parseType(t.Constraint())
	if err != nil {
		return nil, newParseTypeError("parse constraint type", t.Constraint().String(), err)
	}

	return typeParam, nil
}

type parseTypeError struct {
	message    string
	typeString string
	error      error
}

func newParseTypeError(message string, typeString string, error error) *parseTypeError {
	return &parseTypeError{typeString: typeString, error: error, message: message}
}

func (p parseTypeError) Error() string {
	if p.error != nil {
		return fmt.Sprintf("%s: error parsing %s: %s", p.message, p.typeString, p.error)
	}

	return fmt.Sprintf("%s: error parsing type %s", p.message, p.typeString)
}

func (p parseTypeError) Unwrap() error {
	return p.error
}
