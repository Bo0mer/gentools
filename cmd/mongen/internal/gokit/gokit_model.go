package gokit

import (
	"go/ast"

	"github.com/Bo0mer/gentools/cmd/mongen/internal/common"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

type goKitModel struct {
	fileBuilder *astgen.File
	structName  string

	timePackageAlias string
}

func NewGoKitModel(interfacePath, interfaceName, structName, targetPkg string) *goKitModel {
	file := astgen.NewFile(targetPkg)
	strct := astgen.NewStruct(structName)
	file.AppendDeclaration(strct)

	m := &goKitModel{
		fileBuilder: file,
		structName:  structName,
	}
	sourcePackageAlias := m.AddImport("", interfacePath)
	metricsAlias := m.AddImport("", "github.com/go-kit/kit/metrics")
	m.timePackageAlias = m.AddImport("", "time")

	constructorBuilder := newConstructorBuilder(metricsAlias, sourcePackageAlias, interfaceName)
	file.AppendDeclaration(constructorBuilder)

	strct.AddField("next", sourcePackageAlias, interfaceName)
	strct.AddField(common.TotalOpsMetricName, metricsAlias, "Counter")
	strct.AddField(common.FailedOpsMetricName, metricsAlias, "Counter")
	strct.AddField(common.OpsDurationMetricName, metricsAlias, "Histogram")

	return m
}

func (m *goKitModel) AddImport(pkgName, location string) string {
	return m.fileBuilder.AddImport(pkgName, location)
}

func (m *goKitModel) AddMethod(method *astgen.MethodConfig) error {
	mmb := newMonitoringMethodBuilder(m.structName, method)

	mmb.SetTimePackageAlias(m.timePackageAlias)

	m.fileBuilder.AppendDeclaration(mmb)
	return nil
}

func (m *goKitModel) Build() *ast.File {
	return m.fileBuilder.Build()
}

func (m *goKitModel) resolveInterfaceType(location, name string) *ast.SelectorExpr {
	alias := m.AddImport("", location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(alias),
		Sel: ast.NewIdent(name),
	}
}
