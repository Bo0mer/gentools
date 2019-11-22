package main

import "go/ast"

// constructor parameter names
const (
	// metric params
	totalOps    = "totalOps"
	failedOps   = "failedOps"
	opsDuration = "opsDuration"

	// context decorator param
	ctxFuncName = "ctxFunc"
)

// pointerExpr creates a pointer expression for the given package and type. If package is empty it will return just the
// type pointer expression.
func pointerExpr(pkgName, typeName string) ast.Expr {
	typeNameIdent := ast.NewIdent(typeName)
	if pkgName == "" {
		return &ast.StarExpr{
			X: typeNameIdent,
		}
	}
	return &ast.StarExpr{
		X: &ast.SelectorExpr{
			X:   ast.NewIdent(pkgName),
			Sel: typeNameIdent,
		},
	}
}

// buildCtxFuncType builds a FuncType that accepts a context and returns a context.
func buildCtxFuncType(ctxPackageAlias string) *ast.FuncType {
	ctxField := []*ast.Field{
		{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(ctxPackageAlias),
				Sel: ast.NewIdent("Context"),
			},
		},
	}
	return &ast.FuncType{
		Params: &ast.FieldList{
			List: ctxField,
		},
		Results: &ast.FieldList{
			List: ctxField,
		},
	}
}
