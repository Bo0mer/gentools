package generator

import (
	"go/ast"
	"go/token"

	"github.com/mokiat/gostub/util"
)

func NewStubMethodBuilder(methodBuilder *MethodBuilder) *StubMethodBuilder {
	return &StubMethodBuilder{
		methodBuilder: methodBuilder,
		params:        make([]*ast.Field, 0),
		results:       make([]*ast.Field, 0),
	}
}

// StubMethodBuilder is responsible for creating a method that implements
// the original method from the interface and does all the tracking
// logic used by this framework.
//
// Example:
//     func (stub *StubStruct) Sum(a int, b int) int {
//         // ...
//     }
type StubMethodBuilder struct {
	methodBuilder        *MethodBuilder
	mutexFieldSelector   *ast.SelectorExpr
	argsFieldSelector    *ast.SelectorExpr
	returnsFieldSelector *ast.SelectorExpr
	stubFieldSelector    *ast.SelectorExpr
	params               []*ast.Field
	results              []*ast.Field
}

func (b *StubMethodBuilder) SetMutexFieldSelector(selector *ast.SelectorExpr) {
	b.mutexFieldSelector = selector
}

func (b *StubMethodBuilder) SetArgsFieldSelector(selector *ast.SelectorExpr) {
	b.argsFieldSelector = selector
}

func (b *StubMethodBuilder) SetReturnsFieldSelector(selector *ast.SelectorExpr) {
	b.returnsFieldSelector = selector
}

func (b *StubMethodBuilder) SetStubFieldSelector(selector *ast.SelectorExpr) {
	b.stubFieldSelector = selector
}

// SetParams specifies the parameters that the original method
// uses. These parameters need to have been normalized and resolved
// in advance.
func (b *StubMethodBuilder) SetParams(params []*ast.Field) {
	b.params = params
}

// SetResults specifies the results that the original method
// returns. These results need to have been normalized and resolved
// in advance.
func (b *StubMethodBuilder) SetResults(results []*ast.Field) {
	b.results = results
}

func (b *StubMethodBuilder) Build() ast.Decl {
	mutexLockBuilder := NewMutexActionBuilder()
	mutexLockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexLockBuilder.SetAction("Lock")

	mutexUnlockBuilder := NewMutexActionBuilder()
	mutexUnlockBuilder.SetMutexFieldSelector(b.mutexFieldSelector)
	mutexUnlockBuilder.SetAction("Unlock")
	mutexUnlockBuilder.SetDeferred(true)

	b.methodBuilder.SetType(&ast.FuncType{
		Params: &ast.FieldList{
			List: b.params,
		},
		Results: &ast.FieldList{
			List: util.FieldsAsAnonymous(b.results),
		},
	})
	b.methodBuilder.AddStatementBuilder(mutexLockBuilder)
	b.methodBuilder.AddStatementBuilder(mutexUnlockBuilder)

	paramSelectors := []ast.Expr{}
	for _, param := range b.params {
		paramSelectors = append(paramSelectors, ast.NewIdent(param.Names[0].String()))
	}

	b.methodBuilder.AddStatementBuilder(StatementToBuilder(&ast.AssignStmt{
		Lhs: []ast.Expr{
			b.argsFieldSelector,
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: ast.NewIdent("append"),
				Args: []ast.Expr{
					b.argsFieldSelector,
					&ast.CompositeLit{
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: util.FieldsWithoutEllipsis(b.params),
							},
						},
						Elts: paramSelectors,
					},
				},
			},
		},
	}))

	hasEllipsis := false
	if parCount := len(b.params); parCount > 0 {
		if _, ok := b.params[parCount-1].Type.(*ast.Ellipsis); ok {
			hasEllipsis = true
		}
	}

	b.methodBuilder.AddStatementBuilder(StatementToBuilder(&ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  b.stubFieldSelector,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: b.buildCallStubMethodCode(paramSelectors, hasEllipsis),
		Else: b.buildReturnReturnsCode(),
	}))

	return b.methodBuilder.Build()
}

func (b *StubMethodBuilder) buildCallStubMethodCode(args []ast.Expr, hasEllipsis bool) *ast.BlockStmt {
	ellipsisPos := token.NoPos
	if hasEllipsis {
		ellipsisPos = 1
	}
	callExpr := &ast.CallExpr{
		Ellipsis: ellipsisPos,
		Fun:      b.stubFieldSelector,
		Args:     args,
	}
	var stmt ast.Stmt
	if len(b.results) > 0 {
		stmt = &ast.ReturnStmt{
			Results: []ast.Expr{
				callExpr,
			},
		}
	} else {
		stmt = &ast.ExprStmt{
			X: callExpr,
		}
	}
	return &ast.BlockStmt{
		List: []ast.Stmt{
			stmt,
		},
	}
}

func (b *StubMethodBuilder) buildReturnReturnsCode() ast.Stmt {
	if len(b.results) == 0 {
		return nil
	}
	resultSelectors := []ast.Expr{}
	for _, result := range b.results {
		resultSelectors = append(resultSelectors, &ast.SelectorExpr{
			X:   b.returnsFieldSelector,
			Sel: ast.NewIdent(result.Names[0].String()),
		})
	}
	return &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: resultSelectors,
			},
		},
	}
}
