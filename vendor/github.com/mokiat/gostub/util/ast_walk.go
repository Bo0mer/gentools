package util

import "go/ast"

func FieldTypeReuseCount(field *ast.Field) int {
	if len(field.Names) == 0 {
		return 1
	}
	return len(field.Names)
}

func EachFieldInFieldList(fieldList *ast.FieldList) <-chan *ast.Field {
	result := make(chan *ast.Field)
	go func() {
		if fieldList == nil {
			close(result)
			return
		}
		for _, field := range fieldList.List {
			result <- field
		}
		close(result)
	}()
	return result
}

func EachDeclarationInFile(file *ast.File) <-chan ast.Decl {
	result := make(chan ast.Decl)
	go func() {
		for _, decl := range file.Decls {
			result <- decl
		}
		close(result)
	}()
	return result
}

func EachGenericDeclarationInFile(file *ast.File) <-chan *ast.GenDecl {
	result := make(chan *ast.GenDecl)
	go func() {
		for decl := range EachDeclarationInFile(file) {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				result <- genDecl
			}
		}
		close(result)
	}()
	return result
}

func EachSpecificationInGenericDeclaration(decl *ast.GenDecl) <-chan ast.Spec {
	result := make(chan ast.Spec)
	go func() {
		for _, spec := range decl.Specs {
			result <- spec
		}
		close(result)
	}()
	return result
}

func EachTypeSpecificationInGenericDeclaration(decl *ast.GenDecl) <-chan *ast.TypeSpec {
	result := make(chan *ast.TypeSpec)
	go func() {
		for spec := range EachSpecificationInGenericDeclaration(decl) {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				result <- typeSpec
			}
		}
		close(result)
	}()
	return result
}

func EachTypeSpecificationInFile(file *ast.File) <-chan *ast.TypeSpec {
	result := make(chan *ast.TypeSpec)
	go func() {
		for decl := range EachGenericDeclarationInFile(file) {
			for spec := range EachTypeSpecificationInGenericDeclaration(decl) {
				result <- spec
			}
		}
		close(result)
	}()
	return result
}
