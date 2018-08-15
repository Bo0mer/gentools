package resolution

import (
	"go/ast"

	"github.com/Bo0mer/gentools/pkg/internal"
)

type Importer interface {
	AddImport(pkgName, location string) string
}

func NewResolver(importer Importer, locator *Locator) *Resolver {
	return &Resolver{
		importer: importer,
		locator:  locator,
	}
}

type Resolver struct {
	importer Importer
	locator  *Locator
}

func (r *Resolver) ResolveType(context *LocatorContext, astType ast.Expr) (ast.Expr, error) {
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

func (r *Resolver) resolveIdent(context *LocatorContext, ident *ast.Ident) (ast.Expr, error) {
	if r.isBuiltIn(ident.String()) {
		return ident, nil
	}
	discovery, err := r.locator.FindIdentType(context, ident)
	if err != nil {
		return nil, err
	}
	al := r.importer.AddImport("", discovery.Location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(al),
		Sel: ast.NewIdent(ident.String()),
	}, nil
}

func (r *Resolver) resolveSelectorExpr(context *LocatorContext, expr *ast.SelectorExpr) (ast.Expr, error) {
	discovery, err := r.locator.FindSelectorType(context, expr)
	if err != nil {
		return nil, err
	}
	al := r.importer.AddImport("", discovery.Location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(al),
		Sel: expr.Sel,
	}, nil
}

func (r *Resolver) resolveArrayType(context *LocatorContext, astType *ast.ArrayType) (ast.Expr, error) {
	var err error
	astType.Elt, err = r.ResolveType(context, astType.Elt)
	return astType, err
}

func (r *Resolver) resolveMapType(context *LocatorContext, astType *ast.MapType) (ast.Expr, error) {
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

func (r *Resolver) resolveChanType(context *LocatorContext, astType *ast.ChanType) (ast.Expr, error) {
	var err error
	astType.Value, err = r.ResolveType(context, astType.Value)
	return astType, err
}

func (r *Resolver) resolveStarType(context *LocatorContext, astType *ast.StarExpr) (ast.Expr, error) {
	var err error
	astType.X, err = r.ResolveType(context, astType.X)
	return astType, err
}

func (r *Resolver) resolveFuncType(context *LocatorContext, astType *ast.FuncType) (ast.Expr, error) {
	var err error
	for param := range internal.EachFieldInFieldList(astType.Params) {
		param.Type, err = r.ResolveType(context, param.Type)
		if err != nil {
			return nil, err
		}
	}
	for result := range internal.EachFieldInFieldList(astType.Results) {
		result.Type, err = r.ResolveType(context, result.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveStructType(context *LocatorContext, astType *ast.StructType) (ast.Expr, error) {
	var err error
	for field := range internal.EachFieldInFieldList(astType.Fields) {
		field.Type, err = r.ResolveType(context, field.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveInterfaceType(context *LocatorContext, astType *ast.InterfaceType) (ast.Expr, error) {
	var err error
	for field := range internal.EachFieldInFieldList(astType.Methods) {
		field.Type, err = r.ResolveType(context, field.Type)
		if err != nil {
			return nil, err
		}
	}
	return astType, nil
}

func (r *Resolver) resolveEllipsisType(context *LocatorContext, astType *ast.Ellipsis) (ast.Expr, error) {
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
