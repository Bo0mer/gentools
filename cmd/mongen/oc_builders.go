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

// Build builds the constructor method for given monitoring wrapper service using opencensus metrics.
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

	funcParamExpr := func(name, pkg, pkgSel string, asPointer bool) *ast.Field {
		if asPointer {
			return &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(name)},
				Type: &ast.StarExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent(pkg),
						Sel: ast.NewIdent(pkgSel),
					},
				},
			}
		} else {
			return &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(name)},
				Type: &ast.SelectorExpr{
					X:   ast.NewIdent(pkg),
					Sel: ast.NewIdent(pkgSel),
				},
			}
		}
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
					funcParamExpr("next", c.interfacePackageName, c.interfaceName, false),
					funcParamExpr(totalOps, c.metricsPackageName, "Int64Measure", true),
					funcParamExpr(failedOps, c.metricsPackageName, "Int64Measure", true),
					funcParamExpr(opsDuration, c.metricsPackageName, "Float64Measure", true),
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
		tagKeyVarName  = "tagKey"
		ctxFieldName   = "ctx"
	)

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

	snakeCaseMethodName := toSnakeCase(b.methodConfig.MethodName)

	// Create an opencensus tag
	//   tagKey, _ := tag.NewKey("operation")
	createTagKey := &CreateTagKey{
		ctxFieldName:      ctxFieldName,
		tagPackageAlias:   b.packageAliases.tagPkg,
		tagKeyVarName:     tagKeyVarName,
		wrappedMethodName: snakeCaseMethodName,
	}
	b.method.AddStatement(createTagKey.Build())

	//  Insert the tag in to context
	//   ctx, _ = tag.New(ctx, tag.Insert(tagKey, toSnakeCase(c.operationName)))
	insertInContext := InsertTagInContext{
		ctxFieldName:      ctxFieldName,
		tagPackageAlias:   b.packageAliases.tagPkg,
		tagKeyVarName:     tagKeyVarName,
		wrappedMethodName: snakeCaseMethodName,
	}
	b.method.AddStatement(insertInContext.Build())

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

type CreateTagKey struct {
	ctxFieldName      string
	tagPackageAlias   string
	tagKeyVarName     string
	wrappedMethodName string
}

// Build builds a opencensus tag key.
//   tagKey, _ := tag.NewKey("operation")
func (t CreateTagKey) Build() ast.Stmt {
	// tagKey, _ :=
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(t.tagKeyVarName),
			ast.NewIdent("_"),
		},
		Tok: token.DEFINE,
		// ... tag.NewKey("operation")
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(t.tagPackageAlias),
					Sel: ast.NewIdent("NewKey"),
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"operation"`},
				},
			},
		},
	}
}

type InsertTagInContext struct {
	ctxFieldName      string
	tagPackageAlias   string
	tagKeyVarName     string
	wrappedMethodName string
}

// Build creats a new context and adds to it the tag key with the method name as a value.
//   ctx, _ = tag.New(ctx, tag.Insert(tagKey, [t.wrapped_method_name]))
func (t InsertTagInContext) Build() ast.Stmt {
	// tag.Insert(tagKey, [t.wrapped_method_name])
	insertKey := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(t.tagPackageAlias),
			Sel: ast.NewIdent("Insert"),
		},
		Args: []ast.Expr{
			ast.NewIdent(t.tagKeyVarName),
			&ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, t.wrappedMethodName)},
		},
	}

	// ctx, _ = ...
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(t.ctxFieldName),
			ast.NewIdent("_"),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				// ... tag.New(...
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(t.tagPackageAlias),
					Sel: ast.NewIdent("New"),
				},
				/// ... ctx, [tagInsertFuncCall])
				Args: []ast.Expr{
					ast.NewIdent(t.ctxFieldName),
					insertKey,
				},
			},
		},
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
