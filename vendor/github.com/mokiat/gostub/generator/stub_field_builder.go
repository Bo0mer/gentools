package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/util"
)

func NewMethodStubFieldBuilder() *MethodStubFieldBuilder {
	return &MethodStubFieldBuilder{
		params:  make([]*ast.Field, 0),
		results: make([]*ast.Field, 0),
	}
}

// The MethodStubFieldBuilder is responsible for creating the field
// which can be used by end-users to stub the behavior of a given
// method.
//
// Example:
//     type StubStruct struct {
//         // ...
//         SumStub func(a int, b int) (c int)
//         // ...
//     }
type MethodStubFieldBuilder struct {
	fieldName string
	params    []*ast.Field
	results   []*ast.Field
}

func (b *MethodStubFieldBuilder) SetFieldName(name string) {
	b.fieldName = name
}

// SetParams configures the parameters of the stub's function type.
// The parameters should have been normalized and resolved before being
// passed to this function.
func (b *MethodStubFieldBuilder) SetParams(params []*ast.Field) {
	b.params = params
}

// SetResults configures the results of the stub's function type.
// The results should have been normalized and resolved before being
// passed to this function.
func (b *MethodStubFieldBuilder) SetResults(results []*ast.Field) {
	b.results = results
}

func (b *MethodStubFieldBuilder) Build() *ast.Field {
	funcType := &ast.FuncType{
		Params: &ast.FieldList{
			List: b.params,
		},
		Results: &ast.FieldList{
			List: b.results,
		},
	}
	return util.CreateField(b.fieldName, funcType)
}
