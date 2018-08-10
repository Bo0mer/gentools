package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/util"
)

func NewGUIDFieldBuilder() *GUIDFieldBuilder {
	return &GUIDFieldBuilder{}
}

// The GUIDFieldBuilder is responsible for creating a field
// which can be used by end-users to force two stub instances
// not to be equal.
//
// Example:
//     type StubStruct struct {
//         StubGUID int
//         // ...
//     }
type GUIDFieldBuilder struct {
	fieldName string
}

func (b *GUIDFieldBuilder) SetFieldName(name string) {
	b.fieldName = name
}

func (b *GUIDFieldBuilder) Build() *ast.Field {
	return util.CreateField(b.fieldName, ast.NewIdent("int"))
}
