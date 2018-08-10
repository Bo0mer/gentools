package generator

import "go/ast"

func NewMutexActionBuilder() *MutexActionBuilder {
	return &MutexActionBuilder{}
}

// MutexActionBuilder is responsible for creating a statement for
// a mutex of a given stub method.
//
// Example:
//     func (stub *StubStruct) SumCallCount() int {
//         // ...
//         stub.sumMutex.RLock()
//         // ...
//     }
type MutexActionBuilder struct {
	mutexFieldSelector *ast.SelectorExpr
	action             string
	deferred           bool
}

func (b *MutexActionBuilder) SetMutexFieldSelector(selector *ast.SelectorExpr) {
	b.mutexFieldSelector = selector
}

func (b *MutexActionBuilder) SetAction(action string) {
	b.action = action
}

func (b *MutexActionBuilder) SetDeferred(deferred bool) {
	b.deferred = deferred
}

func (b *MutexActionBuilder) Build() ast.Stmt {
	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   b.mutexFieldSelector,
			Sel: ast.NewIdent(b.action),
		},
	}
	if b.deferred {
		return &ast.DeferStmt{
			Call: callExpr,
		}
	} else {
		return &ast.ExprStmt{
			X: callExpr,
		}
	}
}
