package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"unicode"

	"github.com/Bo0mer/gentools/pkg/astgen"
	"github.com/Bo0mer/gentools/pkg/resolution"
)

func parseArgs() (sourceDir, interfaceName string, err error) {
	flag.Parse()
	if flag.NArg() != 2 {
		return "", "", errors.New("too many arguments provided")
	}

	sourceDir = flag.Arg(0)
	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return "", "", fmt.Errorf("error determining absolute path to source directory: %v", err)
	}
	interfaceName = flag.Arg(1)

	return sourceDir, interfaceName, nil
}

func main() {
	sourceDir, interfaceName, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	sourcePkgPath, err := dirToImport(sourceDir)
	if err != nil {
		log.Fatalf("error resolving import path of source directory: %v", err)
	}
	targetPkg := path.Base(sourcePkgPath) + "mws"

	locator := resolution.NewLocator()

	context := resolution.NewSingleLocationContext(sourcePkgPath)
	d, err := locator.FindIdentType(context, ast.NewIdent(interfaceName))
	if err != nil {
		log.Fatal(err)
	}

	typeName := fmt.Sprintf("monitoring%s", interfaceName)

	model := newModel(sourcePkgPath, interfaceName, typeName, targetPkg)
	generator := astgen.Generator{
		Model:    model,
		Locator:  locator,
		Resolver: resolution.NewResolver(model, locator),
	}

	err = generator.ProcessInterface(d)
	if err != nil {
		log.Fatal(err)
	}

	targetPkgPath := filepath.Join(sourceDir, targetPkg)
	if err := os.MkdirAll(targetPkgPath, 0777); err != nil {
		log.Fatalf("error creating target package directory: %v", err)
	}

	fd, err := os.Create(filepath.Join(targetPkgPath, filename(interfaceName)))
	if err != nil {
		log.Fatalf("error creating output source file: %v", err)
	}
	defer fd.Close()

	err = model.WriteSource(fd)
	if err != nil {
		log.Fatal(err)
	}

	wd, _ := os.Getwd()
	path, err := filepath.Rel(wd, fd.Name())
	if err != nil {
		path = fd.Name()
	}
	fmt.Printf("Wrote monitoring implementation of %q to %q\n", sourcePkgPath+"."+interfaceName, path)
}

func filename(interfaceName string) string {
	return fmt.Sprintf("monitoring_%s.go", toSnakeCase(interfaceName))
}

func dirToImport(p string) (string, error) {
	pkg, err := build.ImportDir(p, build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.ImportPath, nil
}

func importToDir(imp string) (string, error) {
	pkg, err := build.Import(imp, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.Dir, nil
}

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
						Names: []*ast.Ident{ast.NewIdent("totalOps")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Counter"),
						},
					},
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent("failedOps")},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(c.metricsPackageName),
							Sel: ast.NewIdent("Counter"),
						},
					},
					&ast.Field{
						Names: []*ast.Ident{ast.NewIdent("opsDuration")},
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
		totalOps:     selexpr("totalOps"),
		failedOps:    selexpr("failedOps"),
		opsDuration:  selexpr("opsDuration"),
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
}

func RecordStartTime(timePackageAlias string) *startTimeRecorder {
	return &startTimeRecorder{timePackageAlias}
}

func (r *startTimeRecorder) Build() ast.Stmt {
	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent(r.timePackageAlias),
			Sel: ast.NewIdent("Now"),
		},
	}

	return &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("_start")},
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

func (r *RecordOpDuration) Build() ast.Stmt {
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

func toSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

func fieldsAsAnonymous(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, len(fields))
	for i, field := range fields {
		result[i] = &ast.Field{
			Type: field.Type,
		}
	}
	return result
}
