package transformation

import (
	"go/ast"
	"unicode"
)

// ToSnakeCase transforms the provided string into snake case by inserting underscores between upper case or number word
// boundaries and lower casing the result.
func ToSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 &&
			(unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) &&
			((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// FieldsAsAnonymous removes the names out of fields and returns only the types. E.g.:
//   (result string, err error) -> (string, error)
//
// It creates and returns a copy of the list and does not modify the provided one.
func FieldsAsAnonymous(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, len(fields))
	for i, field := range fields {
		result[i] = &ast.Field{
			Type: field.Type,
		}
	}
	return result
}
