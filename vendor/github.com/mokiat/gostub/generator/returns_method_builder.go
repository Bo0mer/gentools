package generator

import (
	"go/ast"
	"go/token"
)

func NewReturnsMethodBuilder(methodBuilder *MethodBuilder) *ReturnsMethodBuilder {
	return &ReturnsMethodBuilder{
		methodBuilder: methodBuilder,
		results:       make([]*ast.Field, 0),
	}
}

// ReturnsMethodBuilder is responsible for creating a method on the stub
// structure that allows you to specify the results to be returned by
// default when the stub method is called.
//
// Example:
//     func (stub *StubStruct) AddressReturns(name string, number int) {
//         // ...
//     }
type ReturnsMethodBuilder struct {
	methodBuilder        *MethodBuilder
	mutexFieldSelector   *ast.SelectorExpr
	returnsFieldSelector *ast.SelectorExpr
	results              []*ast.Field
}

func (b *ReturnsMethodBuilder) SetMutexFieldSelector(selector *ast.SelectorExpr) {
	b.mutexFieldSelector = selector
}

func (b *ReturnsMethodBuilder) SetReturnsFieldSelector(selector *ast.SelectorExpr) {
	b.returnsFieldSelector = selector
}

// SetResults specifies the results that the original method
// uses. These results need to have been normalized and resolved
// in advance.
func (b *ReturnsMethodBuilder) SetResults(results []*ast.Field) {
	b.results = results
}

func (b *ReturnsMethodBuilder) Build() ast.Decl {
	mutexLockBuilder := NewMutexActionBuilder()
	mutexLockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexLockBuilder.SetAction("Lock")

	mutexUnlockBuilder := NewMutexActionBuilder()
	mutexUnlockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexUnlockBuilder.SetAction("Unlock")
	mutexUnlockBuilder.SetDeferred(true)

	b.methodBuilder.SetType(&ast.FuncType{
		Params: &ast.FieldList{
			List: b.results,
		},
	})
	b.methodBuilder.AddStatementBuilder(mutexLockBuilder)
	b.methodBuilder.AddStatementBuilder(mutexUnlockBuilder)

	resultSelectors := []ast.Expr{}
	for _, result := range b.results {
		resultSelectors = append(resultSelectors, ast.NewIdent(result.Names[0].String()))
	}
	b.methodBuilder.AddStatementBuilder(StatementToBuilder(&ast.AssignStmt{
		Lhs: []ast.Expr{
			b.returnsFieldSelector,
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: b.results,
					},
				},
				Elts: resultSelectors,
			},
		},
	}))
	return b.methodBuilder.Build()
}
