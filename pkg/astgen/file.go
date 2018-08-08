package astgen

import (
	"fmt"
	"go/ast"
	"go/token"
)

type DeclarationBuilder interface {
	Build() ast.Decl
}

// File describes a single Go source file.
type File struct {
	packageName   string
	importToAlias map[string]string
	aliasToImport map[string]string
	aliasCounter  int
	declarations  []DeclarationBuilder
}

// NewFile returns new empty source file within the specified package.
func NewFile(packageName string) *File {
	return &File{
		packageName:   packageName,
		importToAlias: map[string]string{},
		aliasToImport: map[string]string{},
	}
}

// AddImport assures that the specified package name in the specified
// location will be added as an import and returns the import package alias.
func (f *File) AddImport(packageName, location string) (importAlias string) {
	alias, locationAlreadyRegistered := f.importToAlias[location]
	if locationAlreadyRegistered {
		return alias
	}

	_, aliasAlreadyRegistered := f.aliasToImport[packageName]
	if aliasAlreadyRegistered || packageName == "" {
		alias = f.allocateUniqueAlias()
	} else {
		alias = packageName
	}

	f.importToAlias[location] = alias
	f.aliasToImport[alias] = location
	return alias
}

func (f *File) allocateUniqueAlias() string {
	f.aliasCounter++
	return fmt.Sprintf("alias%d", f.aliasCounter)
}

// Build returns AST representing the file.
func (f *File) Build() *ast.File {
	file := &ast.File{
		Name: ast.NewIdent(f.packageName),
	}

	if len(f.aliasToImport) > 0 {
		importDeclaration := &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: token.Pos(1),
			Specs:  []ast.Spec{},
		}
		for alias, location := range f.aliasToImport {
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

	for _, declaration := range f.declarations {
		file.Decls = append(file.Decls, declaration.Build())
	}

	return file
}

// AppendDeclarations appends the specified decleration to the file.
func (f *File) AppendDeclaration(d DeclarationBuilder) {
	f.declarations = append(f.declarations, d)
}
