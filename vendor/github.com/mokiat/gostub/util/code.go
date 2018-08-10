package util

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"

	"golang.org/x/tools/imports"
)

func CreateSourceCode(astFile *ast.File) ([]byte, error) {
	sourceCode := &bytes.Buffer{}
	err := format.Node(sourceCode, token.NewFileSet(), astFile)
	if err != nil {
		return nil, err
	}
	return sourceCode.Bytes(), nil
}

func FixSourceCodeImports(original []byte) ([]byte, error) {
	return imports.Process("", original, &imports.Options{})
}

func FormatSourceCode(original []byte) ([]byte, error) {
	return format.Source(original)
}

func SaveSourceCode(filePath string, sourceCode []byte) error {
	return ioutil.WriteFile(filePath, sourceCode, 0666)
}
