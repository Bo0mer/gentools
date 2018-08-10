package util_test

import (
	. "github.com/mokiat/gostub/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("String", func() {

	Describe("SnakeCase", func() {
		It("has no effect on empty strings", func() {
			Ω(SnakeCase("")).Should(Equal(""))
		})
		It("works on single letters", func() {
			Ω(SnakeCase("H")).Should(Equal("h"))
		})
		It("has no effect on snake case", func() {
			Ω(SnakeCase("hello_world")).Should(Equal("hello_world"))
		})
		It("works for solo upper case words", func() {
			Ω(SnakeCase("HELLO")).Should(Equal("hello"))
		})
		It("works for upper case letters followed by lower case letters", func() {
			Ω(SnakeCase("HELLOWorld")).Should(Equal("hello_world"))
		})
		It("works for lower case letters followed by upper case letters", func() {
			Ω(SnakeCase("helloWORLD")).Should(Equal("hello_world"))
		})
		It("works for upper case letters followed by digits", func() {
			Ω(SnakeCase("HELLO1234")).Should(Equal("hello_1234"))
		})
		It("works for lower case letters followed by digits", func() {
			Ω(SnakeCase("hello1234")).Should(Equal("hello_1234"))
		})
		It("works for digits followed by upper case letters", func() {
			Ω(SnakeCase("1234HELLO")).Should(Equal("1234_hello"))
		})
		It("works for digits followed by lower case letters", func() {
			Ω(SnakeCase("1234hello")).Should(Equal("1234_hello"))
		})
	})

	Describe("ToPrivate", func() {
		It("has no effect on empty strings", func() {
			Ω(ToPrivate("")).Should(Equal(""))
		})
		It("has no effect on private names", func() {
			Ω(ToPrivate("privateName")).Should(Equal("privateName"))
		})
		It("converts upper camel case to lower camel case", func() {
			Ω(ToPrivate("DoSomething")).Should(Equal("doSomething"))
		})
		It("works on single letters", func() {
			Ω(ToPrivate("U")).Should(Equal("u"))
		})
	})
})
