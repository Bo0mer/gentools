package generator

import "go/ast"

func NewCountMethodBuilder(methodBuilder *MethodBuilder) *CountMethodBuilder {
	return &CountMethodBuilder{
		methodBuilder: methodBuilder,
	}
}

// CountMethodBuilder is responsible for creating a method on the stub
// structure that allows you to check how many times the stubbed method
// was called.
//
// Example:
//     func (stub *StubStruct) SumCallCount() int {
//         // ...
//     }
type CountMethodBuilder struct {
	methodBuilder      *MethodBuilder
	mutexFieldSelector *ast.SelectorExpr
	argsFieldSelector  *ast.SelectorExpr
}

func (b *CountMethodBuilder) SetMutexFieldSelector(selector *ast.SelectorExpr) {
	b.mutexFieldSelector = selector
}

func (b *CountMethodBuilder) SetArgsFieldSelector(selector *ast.SelectorExpr) {
	b.argsFieldSelector = selector
}

func (b *CountMethodBuilder) Build() ast.Decl {
	mutexLockBuilder := NewMutexActionBuilder()
	mutexLockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexLockBuilder.SetAction("RLock")

	mutexUnlockBuilder := NewMutexActionBuilder()
	mutexUnlockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexUnlockBuilder.SetAction("RUnlock")
	mutexUnlockBuilder.SetDeferred(true)

	b.methodBuilder.SetType(&ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent("int"),
				},
			},
		},
	})
	b.methodBuilder.AddStatementBuilder(mutexLockBuilder)
	b.methodBuilder.AddStatementBuilder(mutexUnlockBuilder)
	b.methodBuilder.AddStatementBuilder(StatementToBuilder(&ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("len"),
				Args: []ast.Expr{
					b.argsFieldSelector,
				},
			},
		},
	}))
	return b.methodBuilder.Build()
}
