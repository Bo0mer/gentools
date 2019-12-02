package opencensus

import (
	"go/ast"

	"github.com/Bo0mer/gentools/cmd/mongen/internal/commonbuilders"
	"github.com/Bo0mer/gentools/pkg/astgen"
)

// packageAliases holds the aliases of all imported packages in the generated source file.
type packageAliases struct {
	contextPkg string
	timePkg    string
	statsPkg   string
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
		packageAliases: packageAliases{
			contextPkg: file.AddImport("", "context"),
			timePkg:    file.AddImport("", "time"),
			statsPkg:   file.AddImport("", "go.opencensus.io/stats"),
			tagPkg:     file.AddImport("", "go.opencensus.io/tag"),
		},
	}

	sourcePackageAlias := file.AddImport("", interfacePath)

	strct := astgen.NewStruct(structName)
	strct.AddField("next", sourcePackageAlias, interfaceName)
	strct.AddFieldWithType(commonbuilders.TotalOpsMetricName, pointerExpr(m.packageAliases.statsPkg, "Int64Measure"))
	strct.AddFieldWithType(commonbuilders.FailedOpsMetricName, pointerExpr(m.packageAliases.statsPkg, "Int64Measure"))
	strct.AddFieldWithType(commonbuilders.OpsDurationMetricName, pointerExpr(m.packageAliases.statsPkg, "Float64Measure"))
	strct.AddFieldWithType(commonbuilders.ContextDecoratorFuncName, buildCtxFuncType(m.packageAliases.contextPkg))
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
