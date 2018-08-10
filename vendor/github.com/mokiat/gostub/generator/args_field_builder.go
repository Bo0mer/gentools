package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/util"
)

func NewMethodArgsFieldBuilder() *MethodArgsFieldBuilder {
	return &MethodArgsFieldBuilder{
		params: make([]*ast.Field, 0),
	}
}

// The MethodArgsFieldBuilder is responsible for creating the field
// which is internally used to track the arguments that the end-user
// used to call a given method.
//
// Example:
//     type StubStruct struct {
//         // ...
//         sumArgsForCall []struct {
//             name string,
//             age int,
//         }
//         // ...
//     }
type MethodArgsFieldBuilder struct {
	fieldName string
	params    []*ast.Field
}

func (b *MethodArgsFieldBuilder) SetFieldName(name string) {
	b.fieldName = name
}

// SetParams configures the parameters that the original method has.
// The parameters should have been normalized and resolved beforehand.
func (b *MethodArgsFieldBuilder) SetParams(params []*ast.Field) {
	b.params = params
}

func (b *MethodArgsFieldBuilder) Build() *ast.Field {
	return util.CreateField(b.fieldName, &ast.ArrayType{
		Elt: &ast.StructType{
			Fields: &ast.FieldList{
				List: util.FieldsWithoutEllipsis(b.params),
			},
		},
	})
}
