package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

type constructorBuilder struct {
	metricsPackageName   string
	interfacePackageName string
	interfaceName        string
}

func newConstructorBuilder(metricsPackageName, packageName, interfaceName string) *constructorBuilder {
	return &constructorBuilder{
		metricsPackageName:   metricsPackageName,
		interfacePackageName: packageName,
		interfaceName:        interfaceName,
	}
}

func (c *constructorBuilder) Build() ast.Decl {
	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					// TODO(borshukov): Find a better way to do this.
					ast.NewIdent(fmt.Sprintf("&monitoring%s{next, totalOps, failedOps, opsDuration}", c.interfaceName)),
				},
			},
		},
	}

	funcName := fmt.Sprintf("NewMonitoring%s", c.interfaceName)
	return &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{&ast.Comment{
				Text: fmt.Sprintf("// %s creates new monitoring middleware.", funcName),
			}},
		},
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent("next")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.interfacePackageName),
							Sel: ast.NewIdent(c.interfaceName),
						},
					},
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent(totalOps)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Counter"),
						},
					},
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent(failedOps)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Counter"),
						},
					},
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent(opsDuration)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Histogram"),
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent("")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.interfacePackageName),
							Sel: ast.NewIdent(c.interfaceName),
						},
					},
				},
			},
		},
		Body: funcBody,
	}
}

// MonitoringMethodBuilder is responsible for creating a method that implements
// the original method from the interface and does all the measurement and
// recording logic.
type MonitoringMethodBuilder struct {
	methodConfig *astgen.MethodConfig
	method       *astgen.Method

	totalOps    *ast.SelectorExpr // selector for the struct member
	failedOps   *ast.SelectorExpr // selector for the struct member
	opsDuration *ast.SelectorExpr // selector for the struct member

	timePackageAlias string
}

func NewMonitoringMethodBuilder(structName string, methodConfig *astgen.MethodConfig) *MonitoringMethodBuilder {
	method := astgen.NewMethod(methodConfig.MethodName, "m", structName)

	selexpr := func(fieldName string) *ast.SelectorExpr {
		return &ast.SelectorExpr{
			X:   ast.NewIdent("m"),
			Sel: ast.NewIdent(fieldName),
		}
	}

	return &MonitoringMethodBuilder{
		methodConfig: methodConfig,
		method:       method,
		totalOps:     selexpr(totalOps),
		failedOps:    selexpr(failedOps),
		opsDuration:  selexpr(opsDuration),
	}
}

func (b *MonitoringMethodBuilder) SetTimePackageAlias(alias string) {
	b.timePackageAlias = alias
}

func (b *MonitoringMethodBuilder) Build() ast.Decl {
	b.method.SetType(&ast.FuncType{
		Params: &ast.FieldList{
			List: b.methodConfig.MethodParams,
		},
		Results: &ast.FieldList{
			List: fieldsAsAnonymous(b.methodConfig.MethodResults),
		},
	})

	// Add increase total operations statement
	//   m.totalOps.Add(1)
	increaseTotalOps := &CounterAddAction{counterField: b.totalOps, operationName: b.methodConfig.MethodName}
	b.method.AddStatement(increaseTotalOps.Build())

	// Add statement to capture current time
	//   start := time.Now()
	b.method.AddStatement(RecordStartTime(b.timePackageAlias).Build())

	// Add method invocation:
	//   result1, result2 := m.next.Method(arg1, arg2)
	methodInvocation := NewMethodInvocation(b.methodConfig)
	methodInvocation.SetReceiver(&ast.SelectorExpr{
		X:   ast.NewIdent("m"), // receiver name
		Sel: ast.NewIdent("next"),
	})
	b.method.AddStatement(methodInvocation.Build())

	// Record operation duration
	//   m.opsDuration.Observe(time.Since(start))
	b.method.AddStatement(NewRecordOpDuraton(b.timePackageAlias, b.opsDuration, b.methodConfig.MethodName).Build())

	// Add increase failed operations statement
	//   if err != nil { m.failedOps.Add(1) }
	increaseFailedOps := NewIncreaseFailedOps(b.methodConfig, b.failedOps)
	b.method.AddStatement(increaseFailedOps.Build())

	// Add return statement
	//   return result1, result2
	returnResults := NewReturnResults(b.methodConfig)
	b.method.AddStatement(returnResults.Build())

	return b.method.Build()
}

type CounterAddAction struct {
	counterField  *ast.SelectorExpr
	operationName string
}

func (c *CounterAddAction) Build() ast.Stmt {
	callWithExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   c.counterField,
			Sel: ast.NewIdent("With"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"operation"`},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, toSnakeCase(c.operationName))},
		},
	}

	callAddExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   callWithExpr,
			Sel: ast.NewIdent("Add"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.FLOAT, Value: "1"},
		},
	}

	return &ast.ExprStmt{
		X: callAddExpr,
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

type IncreaseFailedOps struct {
	method       *astgen.MethodConfig
	counterField *ast.SelectorExpr
}

func NewIncreaseFailedOps(m *astgen.MethodConfig, counterField *ast.SelectorExpr) *IncreaseFailedOps {
	return &IncreaseFailedOps{m, counterField}
}

func (i *IncreaseFailedOps) Build() ast.Stmt {
	var errorResult ast.Expr
	for _, result := range i.method.MethodResults {
		if id, ok := result.Type.(*ast.Ident); ok {
			if id.Name == "error" {
				errorResult = ast.NewIdent(result.Names[0].String())
				break
			}
		}
	}

	if errorResult == nil {
		return &ast.EmptyStmt{}
	}

	callWithExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   i.counterField,
			Sel: ast.NewIdent("With"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"operation"`},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, toSnakeCase(i.method.MethodName))},
		},
	}

	callAddExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   callWithExpr,
			Sel: ast.NewIdent("Add"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.FLOAT, Value: "1"},
		},
	}

	callStmt := &ast.ExprStmt{
		X: callAddExpr,
	}

	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  errorResult,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{callStmt},
		},
	}
}

type ReturnResults struct {
	method *astgen.MethodConfig
}

func NewReturnResults(m *astgen.MethodConfig) *ReturnResults {
	return &ReturnResults{m}
}

func (r *ReturnResults) Build() ast.Stmt {
	resultSelectors := []ast.Expr{}
	for _, result := range r.method.MethodResults {
		resultSelectors = append(resultSelectors, ast.NewIdent(result.Names[0].String()))
	}

	return &ast.ReturnStmt{
		Results: resultSelectors,
	}
}

type startTimeRecorder struct {
	timePackageAlias string
	startFieldName   string
}

func RecordStartTime(timePackageAlias string) *startTimeRecorder {
	return &startTimeRecorder{timePackageAlias: timePackageAlias}
}

func (r startTimeRecorder) Build() ast.Stmt {
	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(r.timePackageAlias),
			Sel: ast.NewIdent("Now"),
		},
	}

	startFieldName := r.startFieldName
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

type RecordOpDuration struct {
	timePackageAlias string
	opsDuration      *ast.SelectorExpr
	operationName    string
}

func NewRecordOpDuraton(timePackageAlias string, opsDuration *ast.SelectorExpr, operationName string) *RecordOpDuration {
	return &RecordOpDuration{
		timePackageAlias: timePackageAlias,
		opsDuration:      opsDuration,
		operationName:    operationName,
	}
}

func (r RecordOpDuration) Build() ast.Stmt {
	timeSinceCallExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(r.timePackageAlias),
			Sel: ast.NewIdent("Since"),
		},
		Args: []ast.Expr{ast.NewIdent("_start")},
	}

	durationSecondsExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   timeSinceCallExpr,
			Sel: ast.NewIdent("Seconds"),
		},
	}

	callWithExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   r.opsDuration,
			Sel: ast.NewIdent("With"),
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"operation"`},
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, toSnakeCase(r.operationName))},
		},
	}

	observeCallExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   callWithExpr,
			Sel: ast.NewIdent("Observe"),
		},
		Args: []ast.Expr{durationSecondsExpr},
	}

	return &ast.ExprStmt{X: observeCallExpr}
}
