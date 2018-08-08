package astgen

import (
	"go/ast"
	"go/token"
)

// Struct represents a Go struct.
type Struct struct {
	name   string
	fields []*ast.Field
}

// NewStruct creates new empty struct.
func NewStruct(name string) *Struct {
	return &Struct{
		name: name,
	}
}

// AddField adds field of the specified type to the struct.
func (s *Struct) AddField(name, typePackage, typeName string) {
	s.fields = append(s.fields,
		&ast.Field{
			Names: []*ast.Ident{
				ast.NewIdent(name),
			},
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(typePackage),
				Sel: ast.NewIdent(typeName),
			},
		},
	)
}

func (s *Struct) Build() ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(s.name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: s.fields,
					},
				},
			},
		},
	}
}
