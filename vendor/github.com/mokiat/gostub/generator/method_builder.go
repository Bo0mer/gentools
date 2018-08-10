package generator

import "go/ast"

func NewMethodBuilder() *MethodBuilder {
	return &MethodBuilder{
		statementBuilders: make([]StatementBuilder, 0),
	}
}

type MethodBuilder struct {
	name              string
	funcType          *ast.FuncType
	receiverName      string
	receiverType      string
	statementBuilders []StatementBuilder
}

func (m *MethodBuilder) SetName(name string) {
	m.name = name
}

func (m *MethodBuilder) SetReceiver(name, recType string) {
	m.receiverName = name
	m.receiverType = recType
}

func (m *MethodBuilder) SetType(funcType *ast.FuncType) {
	m.funcType = funcType
}

func (m *MethodBuilder) AddStatementBuilder(builder StatementBuilder) {
	m.statementBuilders = append(m.statementBuilders, builder)
}

func (m *MethodBuilder) Build() ast.Decl {
	statements := make([]ast.Stmt, len(m.statementBuilders))
	for i, builder := range m.statementBuilders {
		statements[i] = builder.Build()
	}
	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{
						ast.NewIdent(m.receiverName),
					},
					Type: &ast.StarExpr{
						X: ast.NewIdent(m.receiverType),
					},
				},
			},
		},
		Name: ast.NewIdent(m.name),
		Type: m.funcType,
		Body: &ast.BlockStmt{
			List: statements,
		},
	}
}
