package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/util"
)

func NewReturnsFieldBuilder() *ReturnsFieldBuilder {
	return &ReturnsFieldBuilder{}
}

// The ReturnsFieldBuilder is responsible for creating the field
// which is used by end-users to specify a stub method's default
// return values.
//
// Example:
//     type StubStruct struct {
//         // ...
//         addressReturns struct {
//             name string,
//             number int,
//         }
//         // ...
//     }
type ReturnsFieldBuilder struct {
	fieldName string
	results   []*ast.Field
}

func (b *ReturnsFieldBuilder) SetFieldName(name string) {
	b.fieldName = name
}

// SetResults configures the results that the original method has.
// The results should have been normalized and resolved beforehand.
func (b *ReturnsFieldBuilder) SetResults(results []*ast.Field) {
	b.results = results
}

func (b *ReturnsFieldBuilder) Build() *ast.Field {
	return util.CreateField(b.fieldName, &ast.StructType{
		Fields: &ast.FieldList{
			List: b.results,
		},
	})
}
