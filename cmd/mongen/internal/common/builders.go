package common

import (
	"go/ast"
	"go/token"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

type StartTimeRecorder struct {
	TimePackageAlias string
	StartFieldName   string
}

func RecordStartTime(timePackageAlias string) *StartTimeRecorder {
	return &StartTimeRecorder{TimePackageAlias: timePackageAlias}
}

// Build builds a statement that records the current timestamp in a new variable.
func (r StartTimeRecorder) Build() ast.Stmt {
	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(r.TimePackageAlias),
			Sel: ast.NewIdent("Now"),
		},
	}

	startFieldName := r.StartFieldName
	if startFieldName == "" {
		startFieldName = "_start"
	}
	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent(startFieldName)},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			callExpr,
		},
	}
}

type MethodInvocation struct {
	receiver *ast.SelectorExpr
	method   *astgen.MethodConfig
}

func (m *MethodInvocation) SetReceiver(s *ast.SelectorExpr) {
	m.receiver = s
}

func NewMethodInvocation(method *astgen.MethodConfig) *MethodInvocation {
	return &MethodInvocation{method: method}
}

// Build builds the call to the original method, as descried in the method configuration.
func (m *MethodInvocation) Build() ast.Stmt {
	resultSelectors := []ast.Expr{}
	for _, result := range m.method.MethodResults {
		resultSelectors = append(resultSelectors, ast.NewIdent(result.Names[0].String()))
	}

	paramSelectors := []ast.Expr{}
	for _, param := range m.method.MethodParams {
		paramSelectors = append(paramSelectors, ast.NewIdent(param.Names[0].String()))
	}

	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   m.receiver,
			Sel: ast.NewIdent(m.method.MethodName),
		},
		Args: paramSelectors,
	}

	if m.method.HasResults() {
		return &ast.AssignStmt{
			Lhs: resultSelectors,
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				callExpr,
			},
		}
	}

	return &ast.ExprStmt{X: callExpr}
}

type ReturnResults struct {
	method *astgen.MethodConfig
}

func NewReturnResults(m *astgen.MethodConfig) *ReturnResults {
	return &ReturnResults{m}
}

// Build builds a return statement based on the method configuration it was created with.
func (r *ReturnResults) Build() ast.Stmt {
	resultSelectors := []ast.Expr{}
	for _, result := range r.method.MethodResults {
		resultSelectors = append(resultSelectors, ast.NewIdent(result.Names[0].String()))
	}

	return &ast.ReturnStmt{
		Results: resultSelectors,
	}
}
