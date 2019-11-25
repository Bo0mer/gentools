package opencensus

import (
	"go/ast"

	"github.com/Bo0mer/gentools/cmd/mongen/internal/common"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

// packageAliases holds the aliases of all imported packages in the generated source file.
type packageAliases struct {
	statsPkg   string
	timePkg    string
	contextPkg string
	tagPkg     string
}

type opencensusModel struct {
	fileBuilder *astgen.File
	structName  string

	packageAliases packageAliases
}

func NewOpencensusModel(interfacePath, interfaceName, structName, targetPkg string) *opencensusModel {
	file := astgen.NewFile(targetPkg)

	m := &opencensusModel{
		fileBuilder: file,
		structName:  structName,
	}

	m.packageAliases = packageAliases{}
	m.packageAliases.contextPkg = m.AddImport("", "context")
	m.packageAliases.statsPkg = m.AddImport("", "go.opencensus.io/stats")
	m.packageAliases.tagPkg = m.AddImport("", "go.opencensus.io/tag")
	m.packageAliases.timePkg = m.AddImport("", "time")
	sourcePackageAlias := m.AddImport("", interfacePath)

	strct := astgen.NewStruct(structName)
	strct.AddField("next", sourcePackageAlias, interfaceName)
	strct.AddFieldWithType(common.TotalOpsMetricName, pointerExpr(m.packageAliases.statsPkg, "Int64Measure"))
	strct.AddFieldWithType(common.FailedOpsMetricName, pointerExpr(m.packageAliases.statsPkg, "Int64Measure"))
	strct.AddFieldWithType(common.OpsDurationMetricName, pointerExpr(m.packageAliases.statsPkg, "Float64Measure"))
	strct.AddFieldWithType(common.ContextDecoratorFuncName, buildCtxFuncType(m.packageAliases.contextPkg))
	file.AppendDeclaration(strct)

	constructorBuilder := newOCConstructorBuilder(
		m.packageAliases.statsPkg, m.packageAliases.contextPkg, sourcePackageAlias, interfaceName)
	file.AppendDeclaration(constructorBuilder)

	return m
}

func (m *opencensusModel) AddImport(pkgName, location string) string {
	return m.fileBuilder.AddImport(pkgName, location)
}

func (m *opencensusModel) AddMethod(method *astgen.MethodConfig) error {
	mmb := newOCMonitoringMethodBuilder(m.structName, method, m.packageAliases)

	m.fileBuilder.AppendDeclaration(mmb)
	return nil
}

func (m *opencensusModel) Build() *ast.File {
	return m.fileBuilder.Build()
}

func (m *opencensusModel) resolveInterfaceType(location, name string) *ast.SelectorExpr {
	alias := m.AddImport("", location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(alias),
		Sel: ast.NewIdent(name),
	}
}
