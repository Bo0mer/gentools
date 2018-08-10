package generator

import (
	"go/ast"
	"go/token"
)

func NewStructBuilder() *StructBuilder {
	return &StructBuilder{
		fieldBuilders: make([]FieldBuilder, 0),
	}
}

type StructBuilder struct {
	name          string
	fieldBuilders []FieldBuilder
}

func (m *StructBuilder) SetName(name string) {
	m.name = name
}

func (m *StructBuilder) AddFieldBuilder(field FieldBuilder) {
	m.fieldBuilders = append(m.fieldBuilders, field)
}

func (m *StructBuilder) Build() ast.Decl {
	fields := make([]*ast.Field, len(m.fieldBuilders))
	for i, builder := range m.fieldBuilders {
		fields[i] = builder.Build()
	}
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(m.name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}
