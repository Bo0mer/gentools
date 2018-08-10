package generator

import (
	"fmt"
	"go/ast"
	"go/token"
)

func NewFileBuilder() *FileBuilder {
	return &FileBuilder{
		importToAlias:              make(map[string]string),
		aliasToImport:              make(map[string]string),
		generalDeclarationBuilders: make([]DeclarationBuilder, 0),
	}
}

type FileBuilder struct {
	filePackageName            string
	importToAlias              map[string]string
	aliasToImport              map[string]string
	aliasCounter               int
	generalDeclarationBuilders []DeclarationBuilder
}

func (m *FileBuilder) SetPackage(name string) {
	m.filePackageName = name
}

// AddImport assures that the specified package name in the specified
// location will be added as an import.
// This function returns the alias to be used in selector expressions.
// If the specified location is already added, then just the alias for
// that package is returned.
func (m *FileBuilder) AddImport(pkgName, location string) string {
	alias, locationAlreadyRegistered := m.importToAlias[location]
	if locationAlreadyRegistered {
		return alias
	}

	_, aliasAlreadyRegistered := m.aliasToImport[pkgName]
	if aliasAlreadyRegistered || pkgName == "" {
		alias = m.allocateUniqueAlias()
	} else {
		alias = pkgName
	}

	m.importToAlias[location] = alias
	m.aliasToImport[alias] = location
	return alias
}

func (m *FileBuilder) allocateUniqueAlias() string {
	m.aliasCounter++
	return fmt.Sprintf("alias%d", m.aliasCounter)
}

func (m *FileBuilder) AddDeclarationBuilder(builder DeclarationBuilder) {
	m.generalDeclarationBuilders = append(m.generalDeclarationBuilders, builder)
}

func (m *FileBuilder) Build() *ast.File {
	file := &ast.File{
		Name: ast.NewIdent(m.filePackageName),
	}

	if len(m.aliasToImport) > 0 {
		importDeclaration := &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: token.Pos(1),
			Specs:  []ast.Spec{},
		}
		for alias, location := range m.aliasToImport {
			importDeclaration.Specs = append(importDeclaration.Specs, &ast.ImportSpec{
				Name: ast.NewIdent(alias),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("\"%s\"", location),
				},
			})
		}
		file.Decls = append(file.Decls, importDeclaration)
	}

	for _, builder := range m.generalDeclarationBuilders {
		file.Decls = append(file.Decls, builder.Build())
	}

	return file
}
