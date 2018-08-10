package generator

import "go/ast"

type DeclarationBuilder interface {
	Build() ast.Decl
}

type FieldBuilder interface {
	Build() *ast.Field
}

type StatementBuilder interface {
	Build() ast.Stmt
}

func StatementToBuilder(statement ast.Stmt) StatementBuilder {
	return &statementBuilder{
		statement: statement,
	}
}

type statementBuilder struct {
	statement ast.Stmt
}

func (b *statementBuilder) Build() ast.Stmt {
	return b.statement
}
