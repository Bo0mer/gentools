package util_test

import (
	"go/ast"

	. "github.com/mokiat/gostub/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AstConstruct", func() {
	Describe("CreateField", func() {
		var field *ast.Field
		var fieldType ast.Expr

		BeforeEach(func() {
			fieldType = ast.NewIdent("string")
			field = CreateField("Name", fieldType)
		})

		It("has correct name", func() {
			Ω(field.Names).ShouldNot(BeNil())
			Ω(field.Names).Should(HaveLen(1))
			Ω(field.Names[0].String()).Should(Equal("Name"))
		})

		It("has correct type", func() {
			Ω(field.Type).Should(Equal(fieldType))
		})
	})

	Describe("FieldsAsAnonymous", func() {
		var fields []*ast.Field

		BeforeEach(func() {
			firstField := &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("Name1"),
					ast.NewIdent("Name2"),
				},
				Type: ast.NewIdent("string"),
			}
			secondField := &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("Name3"),
					ast.NewIdent("Name4"),
				},
				Type: ast.NewIdent("int"),
			}
			fields = []*ast.Field{
				firstField,
				secondField,
			}
		})

		It("returns the same number of fields", func() {
			processed := FieldsAsAnonymous(fields)
			Ω(processed).Should(HaveLen(2))
		})

		It("preserves the types of the fields", func() {
			processed := FieldsAsAnonymous(fields)
			Ω(processed[0].Type).Should(Equal(fields[0].Type))
			Ω(processed[1].Type).Should(Equal(fields[1].Type))
		})

		It("removes any names from the fields", func() {
			processed := FieldsAsAnonymous(fields)
			Ω(processed[0].Names).Should(BeNil())
			Ω(processed[1].Names).Should(BeNil())
		})
	})

	Describe("FieldsWithoutEllipsis", func() {
		var fields []*ast.Field

		BeforeEach(func() {
			firstField := &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("Name1"),
				},
				Type: ast.NewIdent("string"),
			}
			secondField := &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("Name2"),
				},
				Type: &ast.Ellipsis{
					Elt: ast.NewIdent("int"),
				},
			}
			fields = []*ast.Field{
				firstField,
				secondField,
			}
		})

		It("returns the same number of fields", func() {
			processed := FieldsWithoutEllipsis(fields)
			Ω(processed).Should(HaveLen(2))
		})

		It("changes only ellipses types to array types ", func() {
			processed := FieldsWithoutEllipsis(fields)
			Ω(processed[0].Type).Should(Equal(fields[0].Type))
			Ω(processed[1].Type).Should(BeAssignableToTypeOf(&ast.ArrayType{}))
		})
	})
})
