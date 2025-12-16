package main

import (
	"errors"
	"flag"
	"fmt"
	"go/types"
	"strings"

	"go.uber.org/mock/mockgen/model"
	"golang.org/x/tools/go/packages"
)

var (
	buildFlags = flag.String("build_flags", "", "(package mode) Additional flags for go build.")
)

type packageModeParser struct{}

func (p *packageModeParser) parsePackage(packageName string, ifaces []string) (*model.Package, error) {
	pkg, err := p.loadPackage(packageName)
	if err != nil {
		return nil, fmt.Errorf("load package: %w", err)
	}

	modelPackage, err := parseExportFile(packageName, ifaces, pkg.ExportFile)
	if err != nil {
		return nil, fmt.Errorf("extract interfaces from package: %w", err)
	}

	return modelPackage, nil
}

func (p *packageModeParser) loadPackage(packageName string) (*packages.Package, error) {
	var buildFlagsSet []string
	if *buildFlags != "" {
		buildFlagsSet = strings.Split(*buildFlags, " ")
	}

	cfg := &packages.Config{
		Mode:       packages.NeedExportFile,
		BuildFlags: buildFlagsSet,
	}
	pkgs, err := packages.Load(cfg, packageName)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("packages length must be 1: %d", len(pkgs))
	}

	if len(pkgs[0].Errors) > 0 {
		errs := make([]error, len(pkgs[0].Errors))
		for i, err := range pkgs[0].Errors {
			errs[i] = err
		}

		return nil, errors.Join(errs...)
	}

	return pkgs[0], nil
}

func extractInterfacesFromPackageTypes(pkgTypes *types.Package, ifaces []string) ([]*model.Interface, error) {
	// If no interfaces specified, discover all interfaces in the package
	if len(ifaces) == 0  {
		return getAllInterfacesFromPackageTypes(pkgTypes)
	}
	scope := pkgTypes.Scope()
	interfaces := make([]*model.Interface, len(ifaces))
	for i, iface := range ifaces {
		obj := scope.Lookup(iface)
		if obj == nil {
			return nil, fmt.Errorf("interface %s does not exist", iface)
		}

		modelIface, err := parseInterface(obj)
		if err != nil {
			return nil, newParseTypeError("parse interface", obj.Name(), err)
		}

		interfaces[i] = modelIface
	}

	return interfaces, nil
}

// getAllInterfacesFromPackageTypes discovers and returns all exported interfaces in the package.
func getAllInterfacesFromPackageTypes(pkgTypes *types.Package) ([]*model.Interface, error) {
	scope := pkgTypes.Scope()
	names := scope.Names()
	var interfaces []*model.Interface
	for _, name := range names {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}
		// Skip unexported types
		if !obj.Exported() {
			continue
		}
		named, ok := types.Unalias(obj.Type()).(*types.Named)
		if !ok {
			continue
		}
		// Check if the underlying type is an interface
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			continue
		}
		// Skip constraint interfaces (those with type constraints)
		if isConstraint(iface) {
			continue
		}
		modelIface, err := parseInterface(obj)
		if err != nil {
			return nil, newParseTypeError("parse interface", obj.Name(), err)
		}
		interfaces = append(interfaces, modelIface)
	}
	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no interfaces found in package %s", pkgTypes.Path())
	}
	return interfaces, nil
}

func parseInterface(obj types.Object) (*model.Interface, error) {
	named, ok := types.Unalias(obj.Type()).(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %T", obj.Name(), obj.Type().Underlying())
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %T", obj.Name(), obj.Type().Underlying())
	}

	if isConstraint(iface) {
		return nil, fmt.Errorf("interface %s is a constraint", obj.Name())
	}

	methods := make([]*model.Method, iface.NumMethods())
	for i := range iface.NumMethods() {
		method := iface.Method(i)
		typedMethod, ok := method.Type().(*types.Signature)
		if !ok {
			return nil, fmt.Errorf("method %s is not a signature", method.Name())
		}

		modelFunc, err := parseFunc(typedMethod)
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
		typeParam, err := parseConstraint(param)
		if err != nil {
			return nil, newParseTypeError("parse type parameter", param.String(), err)
		}

		typeParams[i] = &model.Parameter{Name: param.Obj().Name(), Type: typeParam}
	}

	return &model.Interface{Name: obj.Name(), Methods: methods, TypeParams: typeParams}, nil
}

func isConstraint(t *types.Interface) bool {
	for i := range t.NumEmbeddeds() {
		embed := t.EmbeddedType(i)
		if _, ok := embed.Underlying().(*types.Interface); !ok {
			return true
		}
	}

	return false
}

func parseType(t types.Type) (model.Type, error) {
	switch t := t.(type) {
	case *types.Array:
		elementType, err := parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse array type", t.Elem().String(), err)
		}
		return &model.ArrayType{Len: int(t.Len()), Type: elementType}, nil
	case *types.Slice:
		elementType, err := parseType(t.Elem())
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

		chanType, err := parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse chan type", t.Elem().String(), err)
		}

		return &model.ChanType{Dir: dir, Type: chanType}, nil
	case *types.Signature:
		sig, err := parseFunc(t)
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
			typedParam, err := parseType(typeParam)
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
		key, err := parseType(t.Key())
		if err != nil {
			return nil, newParseTypeError("parse map key", t.Key().String(), err)
		}
		value, err := parseType(t.Elem())
		if err != nil {
			return nil, newParseTypeError("parse map value", t.Elem().String(), err)
		}

		return &model.MapType{Key: key, Value: value}, nil
	case *types.Pointer:
		valueType, err := parseType(t.Elem())
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

func parseFunc(sig *types.Signature) (*model.FuncType, error) {
	var variadic *model.Parameter
	params := make([]*model.Parameter, 0, sig.Params().Len())
	for i := range sig.Params().Len() {
		param := sig.Params().At(i)

		isVariadicParam := i == sig.Params().Len()-1 && sig.Variadic()
		paramType := param.Type()
		if isVariadicParam {
			sliceType, ok := param.Type().(*types.Slice)
			if !ok {
				return nil, newParseTypeError("variadic parameter is not a slice", param.String(), nil)
			}

			paramType = sliceType.Elem()
		}

		modelParamType, err := parseType(paramType)
		if err != nil {
			return nil, newParseTypeError("parse parameter type", paramType.String(), err)
		}

		modelParameter := &model.Parameter{Type: modelParamType, Name: param.Name()}

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

		resultType, err := parseType(result.Type())
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

func parseConstraint(t *types.TypeParam) (model.Type, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type param")
	}

	typeParam, err := parseType(t.Constraint())
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
