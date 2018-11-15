package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"

	"github.com/Bo0mer/gentools/pkg/astgen"
)

type model struct {
	fileBuilder *astgen.File
	structName  string

	timePackageAlias string
	loggerType       string
}

func newModel(interfacePath, interfaceName, loggerType, structName, targetPkg string) *model {
	file := astgen.NewFile(targetPkg)
	strct := astgen.NewStruct(structName)
	file.AppendDeclaration(strct)

	m := &model{
		fileBuilder: file,
		structName:  structName,
		loggerType:  loggerType,
	}
	sourcePackageAlias := m.AddImport("", interfacePath)

	var logPackageAlias string
	if loggerType == "logrus" {
		logPackageAlias = m.AddImport("", "github.com/sirupsen/logrus")
	} else if loggerType == "go_kit_log" {
		logPackageAlias = m.AddImport("", "github.com/go-kit/kit/log")
	} else if loggerType == "stdlog" {
		logPackageAlias = m.AddImport("", "log")
	}

	constructorBuilder := newConstructorBuilder(logPackageAlias, sourcePackageAlias, interfaceName)
	file.AppendDeclaration(constructorBuilder)

	strct.AddField("next", sourcePackageAlias, interfaceName)
	strct.AddField("logger", logPackageAlias, "Logger")

	return m
}

func (m *model) WriteSource(w io.Writer) error {
	fmt.Fprintf(w, "// Code generated by logen. DO NOT EDIT.\n")
	astFile := m.fileBuilder.Build()

	if err := format.Node(w, token.NewFileSet(), astFile); err != nil {
		return err
	}
	return nil
}

func (m *model) AddImport(pkgName, location string) string {
	return m.fileBuilder.AddImport(pkgName, location)
}

func (m *model) AddMethod(method *astgen.MethodConfig) error {
	mmb := NewLoggingMethodBuilder(m.structName, method, m.loggerType)

	m.fileBuilder.AppendDeclaration(mmb)
	return nil
}

func (m *model) resolveInterfaceType(location, name string) *ast.SelectorExpr {
	alias := m.AddImport("", location)
	return &ast.SelectorExpr{
		X:   ast.NewIdent(alias),
		Sel: ast.NewIdent(name),
	}
}
