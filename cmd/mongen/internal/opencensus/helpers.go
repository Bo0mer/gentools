package opencensus

import "go/ast"

// statsRecordCallExpr prepares the opencensus -> stats.Record(ctx, ... statement, sets the
// provided parameter as second argument and returns the built expression.
func statsRecordCallExpr(statsPackageAlias string, ctxFieldName string, statToRecord ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(statsPackageAlias),
			Sel: ast.NewIdent("Record"),
		},
		Args: []ast.Expr{
			ast.NewIdent(ctxFieldName),
			statToRecord,
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
