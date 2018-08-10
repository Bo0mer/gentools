package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/resolution"
	"github.com/mokiat/gostub/util"
)

func NewResolver(model *GeneratorModel, locator *resolution.Locator) *Resolver {
	return &Resolver{
		model:   model,
		locator: locator,
	}
}

type Resolver struct {
	model   *GeneratorModel
	locator *resolution.Locator
}

func (r *Resolver) ResolveType(context *resolution.LocatorContext, astType ast.Expr) (ast.Expr, error) {
	switch t := astType.(type) {
	case *ast.Ident:
		return r.resolveIdent(context, t)
	case *ast.SelectorExpr:
		return r.resolveSelectorExpr(context, t)
	case *ast.ArrayType:
		return r.resolveArrayType(context, t)
	case *ast.MapType:
		return r.resolveMapType(context, t)
	case *ast.ChanType:
		return r.resolveChanType(context, t)
	case *ast.StarExpr:
		return r.resolveStarType(context, t)
	case *ast.FuncType:
		return r.resolveFuncType(context, t)
	case *ast.StructType:
		return r.resolveStructType(context, t)
	case *ast.InterfaceType:
		return r.resolveInterfaceType(context, t)
	case *ast.Ellipsis:
		return r.resolveEllipsisType(context, t)
	}
	return astType, nil
}

func (r *Resolver) resolveIdent(context *resolution.LocatorContext, ident *ast.Ident) (ast.Expr, error) {
	if r.isBuiltIn(ident.String()) {
		return ident, nil
	}
	discovery, err := r.locator.FindIdentType(context, ident)
	if err != nil {
		return nil, err
	}
	al := r.model.AddImport("", discovery.Location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(al),
		Sel: ast.NewIdent(ident.String()),
	}, nil
}

func (r *Resolver) resolveSelectorExpr(context *resolution.LocatorContext, expr *ast.SelectorExpr) (ast.Expr, error) {
	discovery, err := r.locator.FindSelectorType(context, expr)
	if err != nil {
		return nil, err
	}
	al := r.model.AddImport("", discovery.Location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(al),
		Sel: expr.Sel,
	}, nil
}

func (r *Resolver) resolveArrayType(context *resolution.LocatorContext, astType *ast.ArrayType) (ast.Expr, error) {
	var err error
	astType.Elt, err = r.ResolveType(context, astType.Elt)
	return astType, err
}

func (r *Resolver) resolveMapType(context *resolution.LocatorContext, astType *ast.MapType) (ast.Expr, error) {
	var err error
	astType.Key, err = r.ResolveType(context, astType.Key)
	if err != nil {
		return nil, err
	}
	astType.Value, err = r.ResolveType(context, astType.Value)
	if err != nil {
		return nil, err
	}
	return astType, nil
}

func (r *Resolver) resolveChanType(context *resolution.LocatorContext, astType *ast.ChanType) (ast.Expr, error) {
	var err error
	astType.Value, err = r.ResolveType(context, astType.Value)
	return astType, err
}

func (r *Resolver) resolveStarType(context *resolution.LocatorContext, astType *ast.StarExpr) (ast.Expr, error) {
	var err error
	astType.X, err = r.ResolveType(context, astType.X)
	return astType, err
}

func (r *Resolver) resolveFuncType(context *resolution.LocatorContext, astType *ast.FuncType) (ast.Expr, error) {
	var err error
	for param := range util.EachFieldInFieldList(astType.Params) {
		param.Type, err = r.ResolveType(context, param.Type)
		if err != nil {
			return nil, err
		}
	}
	for result := range util.EachFieldInFieldList(astType.Results) {
		result.Type, err = r.ResolveType(context, result.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveStructType(context *resolution.LocatorContext, astType *ast.StructType) (ast.Expr, error) {
	var err error
	for field := range util.EachFieldInFieldList(astType.Fields) {
		field.Type, err = r.ResolveType(context, field.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveInterfaceType(context *resolution.LocatorContext, astType *ast.InterfaceType) (ast.Expr, error) {
	var err error
	for field := range util.EachFieldInFieldList(astType.Methods) {
		field.Type, err = r.ResolveType(context, field.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveEllipsisType(context *resolution.LocatorContext, astType *ast.Ellipsis) (ast.Expr, error) {
	var err error
	astType.Elt, err = r.ResolveType(context, astType.Elt)
	return astType, err
}

// isBuiltIn should return whether a type, specified by its name,
// is native to the language or not.
func (r *Resolver) isBuiltIn(name string) bool {
	switch name {
	case "bool":
		return true
	case "byte":
		return true
	case "complex64", "complex128":
		return true
	case "error":
		return true
	case "float32", "float64":
		return true
	case "int", "int8", "int16", "int32", "int64":
		return true
	case "rune", "string":
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	case "uintptr":
		return true
	default:
		return false
	}
}
