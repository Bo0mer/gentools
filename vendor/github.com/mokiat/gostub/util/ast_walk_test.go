package util_test

import (
	"go/ast"
	"go/token"

	. "github.com/mokiat/gostub/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AST Walk", func() {
	Describe("FieldTypeReuseCount", func() {
		var field *ast.Field

		Context("when field is standard", func() {
			BeforeEach(func() {
				field = &ast.Field{
					Names: []*ast.Ident{
						ast.NewIdent("first"),
						ast.NewIdent("second"),
					},
				}
			})

			It("returns the correct count", func() {
				Ω(FieldTypeReuseCount(field)).Should(Equal(2))
			})
		})

		Context("when field is anonymous", func() {
			BeforeEach(func() {
				field = &ast.Field{}
			})

			It("returns 1", func() {
				Ω(FieldTypeReuseCount(field)).Should(Equal(1))
			})
		})
	})

	Describe("EachFieldInFieldList", func() {
		var fieldList *ast.FieldList
		var firstParam *ast.Field
		var secondParam *ast.Field

		BeforeEach(func() {
			firstParam = &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("first"),
				},
			}
			secondParam = &ast.Field{
				Names: []*ast.Ident{
					ast.NewIdent("second"),
				},
			}
		})

		Context("when field list is standard", func() {
			BeforeEach(func() {
				fieldList = &ast.FieldList{
					List: []*ast.Field{
						firstParam,
						secondParam,
					},
				}
			})

			It("returns all fields", func() {
				fieldChan := EachFieldInFieldList(fieldList)
				Ω(<-fieldChan).Should(Equal(firstParam))
				Ω(<-fieldChan).Should(Equal(secondParam))
				Eventually(fieldChan).Should(BeClosed())
			})
		})

		Context("when field list is nil", func() {
			BeforeEach(func() {
				fieldList = nil
			})

			It("returns no fields", func() {
				fieldChan := EachFieldInFieldList(fieldList)
				Eventually(fieldChan).Should(BeClosed())
			})
		})

		Context("when List in field list is nil", func() {
			BeforeEach(func() {
				fieldList = &ast.FieldList{
					List: nil,
				}
			})

			It("returns no fields", func() {
				fieldChan := EachFieldInFieldList(fieldList)
				Eventually(fieldChan).Should(BeClosed())
			})
		})
	})

	Describe("EachDeclarationInFile", func() {
		var file *ast.File
		var firstDeclaration ast.Decl
		var secondDeclaration ast.Decl
		var thirdDeclaration ast.Decl

		BeforeEach(func() {
			firstDeclaration = &ast.BadDecl{}
			secondDeclaration = &ast.FuncDecl{}
			thirdDeclaration = &ast.GenDecl{}
			file = &ast.File{
				Decls: []ast.Decl{
					firstDeclaration,
					secondDeclaration,
					thirdDeclaration,
				},
			}
		})

		It("returns all declarations", func() {
			decChan := EachDeclarationInFile(file)
			Ω(<-decChan).Should(Equal(firstDeclaration))
			Ω(<-decChan).Should(Equal(secondDeclaration))
			Ω(<-decChan).Should(Equal(thirdDeclaration))
			Eventually(decChan).Should(BeClosed())
		})
	})

	Describe("EachGenericDeclarationInFile", func() {
		var file *ast.File
		var firstDeclaration ast.Decl
		var secondDeclaration ast.Decl
		var thirdDeclaration ast.Decl
		BeforeEach(func() {
			firstDeclaration = &ast.GenDecl{
				Tok: token.IMPORT,
			}
			secondDeclaration = &ast.FuncDecl{}
			thirdDeclaration = &ast.GenDecl{
				Tok: token.CONST,
			}
			file = &ast.File{
				Decls: []ast.Decl{
					firstDeclaration,
					secondDeclaration,
					thirdDeclaration,
				},
			}
		})

		It("returns only generic declarations", func() {
			decChan := EachGenericDeclarationInFile(file)
			Ω(<-decChan).Should(Equal(firstDeclaration))
			Ω(<-decChan).Should(Equal(thirdDeclaration))
			Eventually(decChan).Should(BeClosed())
		})
	})

	Describe("EachSpecificationInGenericDeclaration", func() {
		var decl *ast.GenDecl
		var firstSpec ast.Spec
		var secondSpec ast.Spec

		BeforeEach(func() {
			firstSpec = &ast.ValueSpec{
				Type: ast.NewIdent("first"),
			}
			secondSpec = &ast.ValueSpec{
				Type: ast.NewIdent("second"),
			}
			decl = &ast.GenDecl{
				Specs: []ast.Spec{
					firstSpec,
					secondSpec,
				},
			}
		})

		It("returns all specifications", func() {
			specChan := EachSpecificationInGenericDeclaration(decl)
			Ω(<-specChan).Should(Equal(firstSpec))
			Ω(<-specChan).Should(Equal(secondSpec))
			Eventually(specChan).Should(BeClosed())
		})
	})

	Describe("EachTypeSpecificationInGenericDeclaration", func() {
		var decl *ast.GenDecl
		var firstSpec ast.Spec
		var secondSpec ast.Spec
		var thirdSpec ast.Spec

		BeforeEach(func() {
			firstSpec = &ast.TypeSpec{
				Name: ast.NewIdent("first"),
			}
			secondSpec = &ast.ValueSpec{
				Type: ast.NewIdent("second"),
			}
			thirdSpec = &ast.TypeSpec{
				Name: ast.NewIdent("third"),
			}
			decl = &ast.GenDecl{
				Specs: []ast.Spec{
					firstSpec,
					secondSpec,
					thirdSpec,
				},
			}
		})

		It("returns all specifications", func() {
			specChan := EachTypeSpecificationInGenericDeclaration(decl)
			Ω(<-specChan).Should(Equal(firstSpec))
			Ω(<-specChan).Should(Equal(thirdSpec))
			Eventually(specChan).Should(BeClosed())
		})
	})

	Describe("EachTypeSpecificationInFile", func() {
		var file *ast.File
		var firstSpec ast.Spec
		var thirdSpec ast.Spec

		BeforeEach(func() {
			firstSpec = &ast.TypeSpec{
				Name: ast.NewIdent("first"),
				Type: &ast.StructType{},
			}
			thirdSpec = &ast.TypeSpec{
				Name: ast.NewIdent("third"),
				Type: &ast.InterfaceType{},
			}
			firstDeclaration := &ast.GenDecl{
				Specs: []ast.Spec{
					firstSpec,
				},
			}
			secondDeclaration := &ast.GenDecl{
				Specs: []ast.Spec{
					&ast.ImportSpec{},
				},
			}
			thirdDeclaration := &ast.GenDecl{
				Specs: []ast.Spec{
					thirdSpec,
				},
			}
			file = &ast.File{
				Decls: []ast.Decl{
					firstDeclaration,
					secondDeclaration,
					thirdDeclaration,
				},
			}
		})

		It("returns all type specifications", func() {
			specChan := EachTypeSpecificationInFile(file)
			Ω(<-specChan).Should(Equal(firstSpec))
			Ω(<-specChan).Should(Equal(thirdSpec))
			Eventually(specChan).Should(BeClosed())
		})
	})
})
