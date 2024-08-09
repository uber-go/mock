package main

import (
	"fmt"
	"go/token"
	"go/types"
	"math/rand"
	"os"

	"go.uber.org/mock/mockgen/model"
	"golang.org/x/tools/go/packages"
)

type importModeParser struct {
	pkgName string
}

func (p *importModeParser) parsePackage(packageName string, ifaces []string) (*model.Package, error) {
	p.pkgName = packageName

	pkg, err := p.loadPackage(packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	interfaces, err := p.extractInterfacesFromPackage(pkg, ifaces)
	if err != nil {
		return nil, fmt.Errorf("failed to extract interfaces from package: %w", err)
	}

	return &model.Package{
		Name:       pkg.Types.Name(),
		PkgPath:    packageName,
		Interfaces: interfaces,
	}, nil
}

func (p *importModeParser) loadPackage(packageName string) (*packages.Package, error) {
	const fakeFileTemplate = `
		package main
		
		import (
			"%s"
		)
	`

	dirName := "mockgen_" + p.generateSalt()
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir for fake package: %w", err)
	}
	defer os.RemoveAll(dirName)

	err = os.WriteFile(dirName+"/main.go", []byte(fmt.Sprintf(fakeFileTemplate, packageName)), 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to write fake main.go: %w", err)
	}

	fileSet := token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedDeps | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedEmbedFiles,
		Dir:  dirName,
		Fset: fileSet,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("packages length must be 1: %d", len(pkgs))
	}

	pkg := pkgs[0]
	searchedPkg, ok := pkg.Imports[packageName]
	if !ok {
		return nil, fmt.Errorf("package %s not found", packageName)
	}

	return searchedPkg, nil
}

func (p *importModeParser) generateSalt() string {
	var numbers = []byte("0123456789")

	result := make([]byte, 8)
	for i := range result {
		result[i] = numbers[rand.Intn(len(numbers))]
	}

	return string(result)
}

func (p *importModeParser) extractInterfacesFromPackage(pkg *packages.Package, ifaces []string) ([]*model.Interface, error) {
	interfaces := make([]*model.Interface, len(ifaces))
	for i, iface := range ifaces {
		obj := pkg.Types.Scope().Lookup(iface)
		if obj == nil {
			return nil, fmt.Errorf("interface %s does not exists", iface)
		}

		modelIface, err := p.parseInterface(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interface: %w", err)
		}

		interfaces[i] = modelIface
	}

	return interfaces, nil
}

func (p *importModeParser) parseInterface(obj types.Object) (*model.Interface, error) {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %s", obj.Name(), obj.Type().Underlying().String())
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface. it is a %s", obj.Name(), obj.Type().Underlying().String())
	}

	iface = iface.Complete()

	if !p.checkInterfaceIsNotConstraint(iface) {
		return nil, fmt.Errorf("interface %s is a constraint", obj.Name())
	}

	methods := make([]*model.Method, iface.NumMethods())
	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		typedMethod, ok := method.Type().(*types.Signature)
		if !ok {
			return nil, fmt.Errorf("method %s is not a signature", method.Name())
		}

		modelFunc, err := p.parseFunc(typedMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to parse method: %w", err)
		}

		methods[i] = &model.Method{
			Name:     method.Name(),
			In:       modelFunc.In,
			Out:      modelFunc.Out,
			Variadic: modelFunc.Variadic,
		}
	}

	var typeParams []*model.Parameter
	if named.TypeParams() != nil {
		typeParams = make([]*model.Parameter, named.TypeParams().Len())
		for i := 0; i < named.TypeParams().Len(); i++ {
			param := named.TypeParams().At(i)
			typeParam, err := p.parseConstraint(param)
			if err != nil {
				return nil, fmt.Errorf("failed to parse type parameter: %w", err)
			}

			typeParams[i] = &model.Parameter{Name: param.Obj().Name(), Type: typeParam}
		}
	}

	return &model.Interface{Name: obj.Name(), Methods: methods, TypeParams: typeParams}, nil
}

func (o *importModeParser) checkInterfaceIsNotConstraint(t *types.Interface) bool {
	for i := 0; i < t.NumEmbeddeds(); i++ {
		embed := t.EmbeddedType(i)
		if _, ok := embed.Underlying().(*types.Interface); !ok {
			return false
		}
	}

	return true
}

func (p *importModeParser) parseType(t types.Type) (model.Type, error) {
	switch t := t.(type) {
	case *types.Array:
		elementType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("failed to parse array type: %w", err)
		}
		return &model.ArrayType{Len: int(t.Len()), Type: elementType}, nil
	case *types.Slice:
		elementType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("failed to parse array type: %w", err)
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
			return nil, fmt.Errorf("failed to parse chan type: %w", err)
		}

		return &model.ChanType{Dir: dir, Type: chanType}, nil
	case *types.Signature:
		sig, err := p.parseFunc(t)
		if err != nil {
			return nil, fmt.Errorf("failed to parse signature: %w", err)
		}

		return sig, nil
	case *types.Named:
		var typeParams *model.TypeParametersType
		if t.TypeArgs() != nil {
			typeParams = &model.TypeParametersType{TypeParameters: make([]model.Type, t.TypeArgs().Len())}
			for i := 0; i < t.TypeArgs().Len(); i++ {
				typeParam := t.TypeArgs().At(i)
				typedParam, err := p.parseType(typeParam)
				if err != nil {
					return nil, fmt.Errorf("failed to parse type parameter: %w", err)
				}

				typeParams.TypeParameters[i] = typedParam
			}
		}

		var pkg string
		if t.Obj().Pkg() != nil {
			pkg = t.Obj().Pkg().Path()
		}

		return &model.NamedType{
			Package:    pkg,
			Type:       t.Obj().Name(),
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
			return nil, fmt.Errorf("failed to parse map key type: %w", err)
		}
		value, err := p.parseType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("failed to parse map value type: %w", err)
		}

		return &model.MapType{Key: key, Value: value}, nil
	case *types.Pointer:
		valueType, err := p.parseType(t.Elem())
		if err != nil {
			return nil, fmt.Errorf("failed to parse pointer type: %w", err)
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

func (p *importModeParser) parseFunc(sig *types.Signature) (*model.FuncType, error) {
	var variadic *model.Parameter
	params := make([]*model.Parameter, 0, sig.Params().Len())
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)

		isVariadicParam := i == sig.Params().Len()-1 && sig.Variadic()
		parseType := param.Type()
		if isVariadicParam {
			sliceType, ok := param.Type().(*types.Slice)
			if !ok {
				return nil, fmt.Errorf("variadic parameter is not a slice")
			}

			parseType = sliceType.Elem()
		}

		paramType, err := p.parseType(parseType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse parameter type: %w", err)
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
	for i := 0; i < sig.Results().Len(); i++ {
		result := sig.Results().At(i)

		resultType, err := p.parseType(result.Type())
		if err != nil {
			return nil, fmt.Errorf("failed to parse result type: %w", err)
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

func (p *importModeParser) parseConstraint(t *types.TypeParam) (model.Type, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type param")
	}

	typeParam, err := p.parseType(t.Constraint())
	if err != nil {
		return nil, fmt.Errorf("failed to parse type parameter: %w", err)
	}

	return typeParam, nil
}
