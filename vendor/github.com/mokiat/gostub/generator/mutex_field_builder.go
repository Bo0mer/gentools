package generator

import (
	"go/ast"

	"github.com/mokiat/gostub/util"
)

func NewMethodMutexFieldBuilder() *MethodMutexFieldBuilder {
	return &MethodMutexFieldBuilder{}
}

// The MethodMutexFieldBuilder is responsible for creating the field
// which is internally used to synchronize access to data related
// to a given method.
//
// Example:
//     type StubStruct struct {
//         // ...
//         sumMutex sync.RWMutex
//         // ...
//     }
type MethodMutexFieldBuilder struct {
	fieldName string
	mutexType ast.Expr
}

func (b *MethodMutexFieldBuilder) SetFieldName(name string) {
	b.fieldName = name
}

// SetMutexType is a way to configure the type of mutex to be used.
// The type should have already been resolved.
func (b *MethodMutexFieldBuilder) SetMutexType(mutexType ast.Expr) {
	b.mutexType = mutexType
}

func (b *MethodMutexFieldBuilder) Build() *ast.Field {
	return util.CreateField(b.fieldName, b.mutexType)
}
