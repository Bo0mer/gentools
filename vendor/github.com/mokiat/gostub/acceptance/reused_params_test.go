package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReusedParams", func() {
	var stub *acceptance_stubs.ReusedParamsStub
	var methodWasCalled bool
	var methodFirstArg string
	var methodSecondArg string

	BeforeEach(func() {
		stub = new(acceptance_stubs.ReusedParamsStub)
		methodWasCalled = false
		methodFirstArg = ""
		methodSecondArg = ""
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(ReusedParams)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to get call count", func() {
		stub.Concat("a", "b")
		stub.Concat("c", "d")
		Ω(stub.ConcatCallCount()).Should(Equal(2))
	})

	It("is possible to stub the behavior", func() {
		stub.ConcatStub = func(first, second string) {
			methodWasCalled = true
			methodFirstArg = first
			methodSecondArg = second
		}
		stub.Concat("Hello", "World")
		Ω(methodWasCalled).Should(BeTrue())
		Ω(methodFirstArg).Should(Equal("Hello"))
		Ω(methodSecondArg).Should(Equal("World"))
	})

	It("is possible to get arguments for call", func() {
		stub.Concat("a", "b")
		stub.Concat("c", "d")
		argFirst, argSecond := stub.ConcatArgsForCall(0)
		Ω(argFirst).Should(Equal("a"))
		Ω(argSecond).Should(Equal("b"))

		argFirst, argSecond = stub.ConcatArgsForCall(1)
		Ω(argFirst).Should(Equal("c"))
		Ω(argSecond).Should(Equal("d"))
	})
})
