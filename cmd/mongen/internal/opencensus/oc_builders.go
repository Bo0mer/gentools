package opencensus

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Bo0mer/gentools/cmd/mongen/internal/commonbuilders"
	"github.com/Bo0mer/gentools/pkg/astgen"
	"github.com/Bo0mer/gentools/pkg/transformation"
)

type ocConstructorBuilder struct {
	metricsPackageName   string
	contextPackageName   string
	interfacePackageName string
	interfaceName        string
}

func newOCConstructorBuilder(
	metricsPackageName, contextPackageName, packageName, interfaceName string) *ocConstructorBuilder {
	return &ocConstructorBuilder{
		metricsPackageName:   metricsPackageName,
		contextPackageName:   contextPackageName,
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
							ast.NewIdent(commonbuilders.TotalOpsMetricName),
							ast.NewIdent(commonbuilders.FailedOpsMetricName),
							ast.NewIdent(commonbuilders.OpsDurationMetricName),
							ast.NewIdent(commonbuilders.ContextDecoratorFuncName),
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
				Type:  pointerExpr(pkg, pkgSel),
			}
		}
		return &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(pkg),
				Sel: ast.NewIdent(pkgSel),
			},
		}
	}

	buildCtxFuncParam := func(fieldName string) *ast.Field {
		return &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(fieldName)},
			Type:  buildCtxFuncType(c.contextPackageName),
		}
	}

	funcName := fmt.Sprintf("NewMonitoring%s", c.interfaceName)
	return &ast.FuncDecl{
		Doc: &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: fmt.Sprintf("// %s creates new monitoring middleware.", funcName),
				},
			},
		},
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					funcParamExpr("next", c.interfacePackageName, c.interfaceName, false),
					funcParamExpr(commonbuilders.TotalOpsMetricName, c.metricsPackageName, "Int64Measure", true),
					funcParamExpr(commonbuilders.FailedOpsMetricName, c.metricsPackageName, "Int64Measure", true),
					funcParamExpr(commonbuilders.OpsDurationMetricName, c.metricsPackageName, "Float64Measure", true),
					buildCtxFuncParam(commonbuilders.ContextDecoratorFuncName),
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
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

// ocMonitoringMethodBuilder is responsible for creating a method that implements
// the original method from the interface and does all the measurement and
// recording logic using opencensus.
type ocMonitoringMethodBuilder struct {
	methodConfig *astgen.MethodConfig
	method       *astgen.Method
	receiverName string

	// selectors for the struct members
	totalOps    *ast.SelectorExpr
	failedOps   *ast.SelectorExpr
	opsDuration *ast.SelectorExpr
	ctxFuncSel  *ast.SelectorExpr

	packageAliases packageAliases
}

func newOCMonitoringMethodBuilder(structName string, methodConfig *astgen.MethodConfig, aliases packageAliases) *ocMonitoringMethodBuilder {
	receiverName := "m"
	method := astgen.NewMethod(methodConfig.MethodName, receiverName, structName)

	selexpr := func(fieldName string) *ast.SelectorExpr {
		return &ast.SelectorExpr{
			X:   ast.NewIdent(receiverName),
			Sel: ast.NewIdent(fieldName),
		}
	}

	return &ocMonitoringMethodBuilder{
		methodConfig:   methodConfig,
		method:         method,
		receiverName:   receiverName,
		totalOps:       selexpr(commonbuilders.TotalOpsMetricName),
		failedOps:      selexpr(commonbuilders.FailedOpsMetricName),
		opsDuration:    selexpr(commonbuilders.OpsDurationMetricName),
		ctxFuncSel:     selexpr(commonbuilders.ContextDecoratorFuncName),
		packageAliases: aliases,
	}
}

func (b *ocMonitoringMethodBuilder) Build() ast.Decl {
	// Add the func declaration
	//   func ([b.method.receiverName] [b.method.receiverType]) [funcName]([MethodParams...]) ([MethodResults...]) {
	b.method.SetType(&ast.FuncType{
		Params: &ast.FieldList{
			List: b.methodConfig.MethodParams,
		},
		Results: &ast.FieldList{
			List: transformation.FieldsAsAnonymous(b.methodConfig.MethodResults),
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
	initContextVar := &contextParam{
		ctxFieldName:    ctxFieldName,
		ctxPackageAlias: b.packageAliases.contextPkg,
		methodConfig:    b.methodConfig,
	}
	b.method.AddStatement(initContextVar.Build())

	// Run the context decorator func if provided
	// if m.ctxFunc != nil {
	//   ctx = m.ctxFunc(ctx)
	// }
	ctxDecorator := &contextDecorator{
		ctxFieldName: ctxFieldName,
		ctxFuncSel:   b.ctxFuncSel,
	}
	b.method.AddStatement(ctxDecorator.Build())

	snakeCaseMethodName := transformation.ToSnakeCase(b.methodConfig.MethodName)
	// Create an opencensus tag
	//   tagKey, _ := tag.MustNewKey("operation")
	createTagKey := &createTagKey{
		ctxFieldName:      ctxFieldName,
		tagPackageAlias:   b.packageAliases.tagPkg,
		tagKeyVarName:     tagKeyVarName,
		wrappedMethodName: snakeCaseMethodName,
	}
	b.method.AddStatement(createTagKey.Build())

	//  Insert the tag in to context
	//   var err error
	//   ctx, err = tag.New(ctx, tag.Insert(tagKey, toSnakeCase(c.operationName)))
	//   if err != nil {
	//     panic(err)
	//   }
	insertInContext := insertTagInContext{
		ctxFieldName:      ctxFieldName,
		tagPackageAlias:   b.packageAliases.tagPkg,
		tagKeyVarName:     tagKeyVarName,
		wrappedMethodName: snakeCaseMethodName,
	}
	b.method.AddStatements(insertInContext.Build())

	// Add increase total operations statement
	// 	 stats.Record(ctx, m.totalOps.M(1))
	increaseTotalOps := &recordStat{
		statsPackageAlias: b.packageAliases.statsPkg,
		ctxFieldName:      ctxFieldName,
		statField:         b.totalOps,
	}
	b.method.AddStatement(increaseTotalOps.Build())

	// Add statement to capture current time
	//   start := time.Now()
	b.method.AddStatement(commonbuilders.StartTimeRecorder{
		TimePackageAlias: b.packageAliases.timePkg,
		StartFieldName:   startFieldName,
	}.Build())

	// Add method invocation:
	//   result1, result2 := m.next.Method(arg1, arg2)
	methodInvocation := commonbuilders.NewMethodInvocation(b.methodConfig)
	methodInvocation.SetReceiver(&ast.SelectorExpr{
		X:   ast.NewIdent(b.receiverName),
		Sel: ast.NewIdent("next"),
	})
	b.method.AddStatement(methodInvocation.Build())

	// Record operation duration
	//   stats.Record(ctx, m.opsDuration.M(time.Since(start).Seconds()))
	b.method.AddStatement(recordOpsDurationStats{
		opsDurationField:  b.opsDuration,
		statsPackageAlias: b.packageAliases.statsPkg,
		startFieldName:    startFieldName,
		ctxFieldName:      ctxFieldName,
		timePackageAlias:  b.packageAliases.timePkg,
	}.Build())

	// Add increase failed operations statement
	//   if err != nil { m.failedOps.Add(1) }
	b.method.AddStatement(incrementFailedOps{
		failedOpsField:    b.failedOps,
		method:            b.methodConfig,
		counterField:      "failedOps",
		ctxFieldName:      ctxFieldName,
		statsPackageAlias: b.packageAliases.statsPkg,
	}.Build())

	// Add return statement
	//   return result1, result2
	returnResults := commonbuilders.NewReturnResults(b.methodConfig)
	b.method.AddStatement(returnResults.Build())

	return b.method.Build()
}

type contextParam struct {
	ctxFieldName    string
	ctxPackageAlias string
	methodConfig    *astgen.MethodConfig
}

// Build builds a context variable initialization or assignment. If the methodConfig shows that the first parameter is
// of type context it will be assigned to a a "ctxFieldName" variable. If not ctx will be initialized as context.Background()
// ctx
func (c contextParam) Build() ast.Stmt {
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

type contextDecorator struct {
	ctxFieldName string
	ctxFuncSel   *ast.SelectorExpr
}

// Build builds a statement to decorate the context with the context func if it is provided.
//
//   if m.ctxFunc != nil {
//     ctx = m.ctxFunc(ctx)
//   }
func (c contextDecorator) Build() ast.Stmt {
	ctxSel := ast.NewIdent(c.ctxFieldName)

	// ctx = m.ctxFunc(ctx)
	decorateFuncStmt := &ast.AssignStmt{
		Lhs: []ast.Expr{ctxSel},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun:  c.ctxFuncSel,
				Args: []ast.Expr{ctxSel},
			},
		},
	}

	return &ast.IfStmt{
		// if m.ctxFunc != nil
		Cond: &ast.BinaryExpr{
			X:  c.ctxFuncSel,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{decorateFuncStmt},
		},
	}
}

type createTagKey struct {
	ctxFieldName      string
	tagPackageAlias   string
	tagKeyVarName     string
	wrappedMethodName string
}

// Build builds a opencensus tag key.
//   tagKey, _ := tag.NewKey("operation")
func (t createTagKey) Build() ast.Stmt {
	// tagKey, _ :=
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(t.tagKeyVarName),
		},
		Tok: token.DEFINE,
		// ... tag.NewKey("operation")
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(t.tagPackageAlias),
					Sel: ast.NewIdent("MustNewKey"),
				},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"operation"`},
				},
			},
		},
	}
}

type insertTagInContext struct {
	ctxFieldName      string
	tagPackageAlias   string
	tagKeyVarName     string
	wrappedMethodName string
}

// Build creates a new context and adds to it the tag key with the method name as a value.
//   var err error
//   ctx, _ = tag.New(ctx, tag.Insert(tagKey, toSnakeCase([t.wrapped_method_name])))
//   if err != nil {
//     panic(err)
//   }
func (t insertTagInContext) Build() []ast.Stmt {
	var stmts []ast.Stmt

	errSel := ast.NewIdent("err")
	newTagStmt := t.buildNewTagStmt(errSel)

	stmts = append(stmts, t.buildVarErrStmt(errSel))
	stmts = append(stmts, t.buildIfErrThenPanicStmt(newTagStmt, errSel))

	return stmts
}

// buildVarErrStmt builds the `var err error` statement
func (t insertTagInContext) buildVarErrStmt(errSel *ast.Ident) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: errSel,
					Type: ast.NewIdent("error"),
				},
			},
		},
	}
}

// buildIfErrThenPanicStmt builds the if statement with the error check and panic if it's not nil
//   if [initStmt]; [errSel] != nil {
//     panic([errSel])
//   }
func (t insertTagInContext) buildIfErrThenPanicStmt(initStmt ast.Stmt, errSel ast.Expr) ast.Stmt {
	panicCallExpr := &ast.CallExpr{
		Fun:  ast.NewIdent("panic"),
		Args: []ast.Expr{errSel},
	}

	return &ast.IfStmt{
		Init: initStmt,
		Cond: &ast.BinaryExpr{
			X:  errSel,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					X: panicCallExpr,
				},
			},
		},
	}
}

// buildNewTagStmt builds the creation of the new tag and the assignment to the result variables.
//   [t.ctxFieldName], [errSel] = tag.Insert(tagKey, [t.wrapped_method_name])
func (t insertTagInContext) buildNewTagStmt(errSel ast.Expr) ast.Stmt {
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

	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(t.ctxFieldName),
			errSel,
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

type recordStat struct {
	statField         *ast.SelectorExpr
	ctxFieldName      string
	statsPackageAlias string
}

// Build builds a statement in the form:
// stats.Record(ctx, [statField].M(1))
func (r recordStat) Build() ast.Stmt {
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

type recordOpsDurationStats struct {
	opsDurationField  *ast.SelectorExpr
	startFieldName    string
	statsPackageAlias string
	ctxFieldName      string
	timePackageAlias  string
}

// Build builds a statement in the form:
// stats.Record(ctx, [opsDurationField].M([timePackageAlias].Since([startFieldName]).Seconds()))
func (r recordOpsDurationStats) Build() ast.Stmt {
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

type incrementFailedOps struct {
	failedOpsField    *ast.SelectorExpr
	method            *astgen.MethodConfig
	counterField      string
	ctxFieldName      string
	statsPackageAlias string
}

func (i incrementFailedOps) Build() ast.Stmt {
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
			List: []ast.Stmt{recordStat{
				statsPackageAlias: i.statsPackageAlias,
				statField:         i.failedOpsField,
				ctxFieldName:      i.ctxFieldName,
			}.Build()},
		},
	}
}
