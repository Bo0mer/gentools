package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

type ocConstructorBuilder struct {
	metricsPackageName   string
	interfacePackageName string
	interfaceName        string
}

func newOCConstructorBuilder(metricsPackageName, packageName, interfaceName string) *ocConstructorBuilder {
	return &ocConstructorBuilder{
		metricsPackageName:   metricsPackageName,
		interfacePackageName: packageName,
		interfaceName:        interfaceName,
	}
}

// Build builds the constructor method for given monitoring wrapper service using opensensus metrics.
func (c *ocConstructorBuilder) Build() ast.Decl {
	funcBody := &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CompositeLit{
						Type: ast.NewIdent(fmt.Sprintf("&monitoring%s", c.interfaceName)),
						Elts: []ast.Expr{
							ast.NewIdent("next"),
							ast.NewIdent(totalOps),
							ast.NewIdent(failedOps),
							ast.NewIdent(opsDuration),
						},
					},
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
					{
						Names: []*ast.Ident{ast.NewIdent("next")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.interfacePackageName),
							Sel: ast.NewIdent(c.interfaceName),
						},
					},
					{
						Names: []*ast.Ident{ast.NewIdent(totalOps)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Int64Measure"),
						},
					},
					{
						Names: []*ast.Ident{ast.NewIdent(failedOps)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Int64Measure"),
						},
					},
					{
						Names: []*ast.Ident{ast.NewIdent(opsDuration)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Float64Measure"),
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

// OCMonitoringMethodBuilder is responsible for creating a method that implements
// the original method from the interface and does all the measurement and
// recording logic using opencensus.
type OCMonitoringMethodBuilder struct {
	methodConfig *astgen.MethodConfig
	method       *astgen.Method
	receiverName string

	totalOps    *ast.SelectorExpr // selector for the struct member
	failedOps   *ast.SelectorExpr // selector for the struct member
	opsDuration *ast.SelectorExpr // selector for the struct member

	packageAliases packageAliases
}

func NewOCMonitoringMethodBuilder(structName string, methodConfig *astgen.MethodConfig, aliases packageAliases) *OCMonitoringMethodBuilder {
	receiverName := "m"
	method := astgen.NewMethod(methodConfig.MethodName, receiverName, structName)

	selexpr := func(fieldName string) *ast.SelectorExpr {
		return &ast.SelectorExpr{
			X:   ast.NewIdent(receiverName),
			Sel: ast.NewIdent(fieldName),
		}
	}

	return &OCMonitoringMethodBuilder{
		methodConfig:   methodConfig,
		method:         method,
		receiverName:   receiverName,
		totalOps:       selexpr(totalOps),
		failedOps:      selexpr(failedOps),
		opsDuration:    selexpr(opsDuration),
		packageAliases: aliases,
	}
}

func (b *OCMonitoringMethodBuilder) Build() ast.Decl {
	// Add the func declaration
	//   func ([b.method.receiverName] [b.method.receiverType]) [funcName]([MethodParams...]) ([MethodResults...]) {
	b.method.SetType(&ast.FuncType{
		Params: &ast.FieldList{
			List: b.methodConfig.MethodParams,
		},
		Results: &ast.FieldList{
			List: fieldsAsAnonymous(b.methodConfig.MethodResults),
		},
	})

	const (
		startFieldName = "start"
		ctxFieldName   = "ctx"
	)

	// TODO - do we want to enrich the context with tags? They seem similar to go-kit's labels
	// Add ctx initialization. Can be either
	//   ctx := [contextPkg].Background()
	// or
	//   ctx := [ctxFieldName]
	initContextVar := &ContextParam{
		ctxFieldName:    ctxFieldName,
		ctxPackageAlias: b.packageAliases.contextPkg,
		methodConfig:    b.methodConfig,
	}
	b.method.AddStatement(initContextVar.Build())

	// Add increase total operations statement
	// 	 stats.Record(ctx, m.totalOps.M(1))
	increaseTotalOps := &RecordStat{
		statsPackageAlias: b.packageAliases.statsPkg,
		ctxFieldName:      ctxFieldName,
		statField:         b.totalOps,
	}
	b.method.AddStatement(increaseTotalOps.Build())

	// Add statement to capture current time
	//   start := time.Now()
	b.method.AddStatement(startTimeRecorder{
		timePackageAlias: b.packageAliases.timePkg,
		startFieldName:   startFieldName,
	}.Build())

	// Add method invocation:
	//   result1, result2 := m.next.Method(arg1, arg2)
	methodInvocation := NewMethodInvocation(b.methodConfig)
	methodInvocation.SetReceiver(&ast.SelectorExpr{
		X:   ast.NewIdent(b.receiverName),
		Sel: ast.NewIdent("next"),
	})
	b.method.AddStatement(methodInvocation.Build())

	// Record operation duration
	//   m.opsDuration.Observe(time.Since(start))
	//b.method.AddStatement(NewRecordOpDuraton(b.timePackageAlias, b.opsDuration, b.methodConfig.MethodName).Build())

	// Record operation duration
	//   stats.Record(ctx, m.opsDuration.M(time.Since(start).Seconds()))
	b.method.AddStatement(RecordOpsDurationStats{
		opsDurationField:  b.opsDuration,
		statsPackageAlias: b.packageAliases.statsPkg,
		startFieldName:    startFieldName,
		ctxFieldName:      ctxFieldName,
		timePackageAlias:  b.packageAliases.timePkg,
	}.Build())

	// Add increase failed operations statement
	//   if err != nil { m.failedOps.Add(1) }
	b.method.AddStatement(IncrementFailedOps{
		failedOpsField:    b.failedOps,
		method:            b.methodConfig,
		counterField:      "failedOps",
		ctxFieldName:      ctxFieldName,
		statsPackageAlias: b.packageAliases.statsPkg,
	}.Build())

	// Add return statement
	//   return result1, result2
	returnResults := NewReturnResults(b.methodConfig)
	b.method.AddStatement(returnResults.Build())

	return b.method.Build()
}

type ContextParam struct {
	ctxFieldName    string
	ctxPackageAlias string
	methodConfig    *astgen.MethodConfig
}

// Build builds a context variable initialization or assignment. If the methodConfig shows that the first parameter is
// of type context it will be assigned to a a "ctxFieldName" variable. If not ctx will be initialized as context.Background()
// ctx
func (c ContextParam) Build() ast.Stmt {
	// [c.ctxFieldName] :=
	lhs := []ast.Expr{ast.NewIdent(c.ctxFieldName)}

	// [ctxPackageAlias].Background()
	rhs := []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent(c.ctxPackageAlias),
				Sel: ast.NewIdent("Background"),
			},
		},
	}

	if len(c.methodConfig.MethodParams) > 0 {
		p1 := c.methodConfig.MethodParams[0]
		if sel, ok := p1.Type.(*ast.SelectorExpr); ok {
			if sel.Sel.String() == "Context" {
				if id, ok := sel.X.(*ast.Ident); ok && id.String() == c.ctxPackageAlias {
					// names is already populated by
					rhs = []ast.Expr{p1.Names[0]}
				}
			}
		}
	}

	return &ast.AssignStmt{
		Lhs: lhs,
		Tok: token.DEFINE,
		Rhs: rhs,
	}
}

type RecordStat struct {
	statField         *ast.SelectorExpr
	ctxFieldName      string
	statsPackageAlias string
}

// Build builds a statement in the form:
// stats.Record(ctx, [statField].M(1))
func (r RecordStat) Build() ast.Stmt {
	return &ast.ExprStmt{
		X: statsRecordCallExpr(r.statsPackageAlias, r.ctxFieldName, &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   r.statField,
				Sel: ast.NewIdent("M"),
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: "1"},
			},
		}),
	}
}

type RecordOpsDurationStats struct {
	opsDurationField  *ast.SelectorExpr
	startFieldName    string
	statsPackageAlias string
	ctxFieldName      string
	timePackageAlias  string
}

// Build builds a statement in the form:
// stats.Record(ctx, [opsDurationField].M([timePackageAlias].Since([startFieldName]).Seconds()))
func (r RecordOpsDurationStats) Build() ast.Stmt {
	return &ast.ExprStmt{
		X: statsRecordCallExpr(r.statsPackageAlias, r.ctxFieldName, &ast.CallExpr{
			// stats.Record(ctx, [opsDurationField].M(...)
			Fun: &ast.SelectorExpr{
				X:   r.opsDurationField,
				Sel: ast.NewIdent("M"),
			},
			Args: []ast.Expr{
				// [timePackageAlias].Since([startFieldName]).Seconds()
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent(r.timePackageAlias),
								Sel: ast.NewIdent("Since"),
							},
							Args: []ast.Expr{ast.NewIdent(r.startFieldName)},
						},
						Sel: ast.NewIdent("Seconds"),
					},
				},
			},
		}),
	}
}

type IncrementFailedOps struct {
	failedOpsField    *ast.SelectorExpr
	method            *astgen.MethodConfig
	counterField      string
	ctxFieldName      string
	statsPackageAlias string
}

func (i IncrementFailedOps) Build() ast.Stmt {
	var errorResult ast.Expr
	for _, result := range i.method.MethodResults {
		if id, ok := result.Type.(*ast.Ident); ok {
			if id.Name == "error" {
				errorResult = ast.NewIdent(result.Names[0].String())
				break
			}
		}
	}

	// none of the returns is an error type
	if errorResult == nil {
		return &ast.EmptyStmt{}
	}

	return &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  errorResult,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{RecordStat{
				statsPackageAlias: i.statsPackageAlias,
				statField:         i.failedOpsField,
				ctxFieldName:      i.ctxFieldName,
			}.Build()},
		},
	}
}

// statsRecordCallExpr prepares the opencensus -> stats.Record(ctx, ... statement, sets the
// provided parameter as second argument and returns the built expression.
func statsRecordCallExpr(statsPackageAlias string, ctxFieldName string, statToRecord ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(statsPackageAlias),
			Sel: ast.NewIdent("Record"),
		},
		Args: []ast.Expr{
			ast.NewIdent(ctxFieldName),
			statToRecord,
		},
	}
}
