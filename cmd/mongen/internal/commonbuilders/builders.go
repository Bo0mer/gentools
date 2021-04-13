package commonbuilders

import (
	"go/ast"
	"go/token"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

// constructor parameter names
const (
	// metric params
	TotalOpsMetricName    = "totalOps"
	FailedOpsMetricName   = "failedOps"
	OpsDurationMetricName = "opsDuration"

	// context decorator param
	ContextDecoratorFuncName = "ctxFunc"
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

// TODO: Move MethodInvocation to a reusable package as
// the same implementation can be seen multiple times
// within this project.

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
	ellipsisPos := token.NoPos
	for _, param := range m.method.MethodParams {
		paramSelectors = append(paramSelectors, ast.NewIdent(param.Names[0].String()))
		if p, ok := param.Type.(*ast.Ellipsis); ok {
			ellipsisPos = p.Pos()
		}
	}

	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   m.receiver,
			Sel: ast.NewIdent(m.method.MethodName),
		},
		Args:     paramSelectors,
		Ellipsis: ellipsisPos,
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
