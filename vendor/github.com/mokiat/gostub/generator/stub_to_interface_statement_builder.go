package generator

import (
	"go/ast"
	"go/token"
)

func NewStubToInterfaceStatementBuilder() *StubToInterfaceStatementBuilder {
	return &StubToInterfaceStatementBuilder{}
}

type StubToInterfaceStatementBuilder struct {
	stubName          string
	interfaceSelector *ast.SelectorExpr
}

func (b *StubToInterfaceStatementBuilder) SetStubName(name string) {
	b.stubName = name
}

func (b *StubToInterfaceStatementBuilder) SetInterfaceSelector(selector *ast.SelectorExpr) {
	b.interfaceSelector = selector
}

func (b *StubToInterfaceStatementBuilder) Build() ast.Decl {
	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					ast.NewIdent("_"),
				},
				Type: b.interfaceSelector,
				Values: []ast.Expr{
					&ast.CallExpr{
						Fun: ast.NewIdent("new"),
						Args: []ast.Expr{
							ast.NewIdent(b.stubName),
						},
					},
				},
			},
		},
	}
}
