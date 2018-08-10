package astgen

import "go/ast"

func NewMethod(name, receiverName, receiverType string) *Method {
	return &Method{
		name:         name,
		receiverName: receiverName,
		receiverType: receiverType,
	}
}

// Method represents a struct method.
type Method struct {
	name         string
	receiverName string
	receiverType string
	funcType     *ast.FuncType
	statements   []ast.Stmt
}

func (m *Method) SetType(funcType *ast.FuncType) {
	m.funcType = funcType
}

func (m *Method) AddStatement(stmt ast.Stmt) {
	m.statements = append(m.statements, stmt)
}

func (m *Method) Build() ast.Decl {
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
			List: m.statements,
		},
	}
}
