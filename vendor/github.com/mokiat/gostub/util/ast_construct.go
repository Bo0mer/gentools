package util

import "go/ast"

func CreateField(name string, fieldType ast.Expr) *ast.Field {
	return &ast.Field{
		Names: []*ast.Ident{
			ast.NewIdent(name),
		},
		Type: fieldType,
	}
}

func FieldsAsAnonymous(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, len(fields))
	for i, field := range fields {
		result[i] = &ast.Field{
			Type: field.Type,
		}
	}
	return result
}

func FieldsWithoutEllipsis(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, len(fields))
	for i, field := range fields {
		result[i] = &ast.Field{
			Names: field.Names,
			Type:  field.Type,
		}
		if ellipsisType, ok := field.Type.(*ast.Ellipsis); ok {
			result[i].Type = &ast.ArrayType{
				Elt: ellipsisType.Elt,
			}
		}
	}
	return result
}
