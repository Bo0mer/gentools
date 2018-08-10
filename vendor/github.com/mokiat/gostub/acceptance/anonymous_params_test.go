package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnonymousParams", func() {
	var stub *acceptance_stubs.AnonymousParamsStub
	var methodWasCalled bool
	var methodFirstArg string
	var methodSecondArg int

	BeforeEach(func() {
		stub = new(acceptance_stubs.AnonymousParamsStub)
		methodWasCalled = false
		methodFirstArg = ""
		methodSecondArg = 0
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(AnonymousParams)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to get call count", func() {
		stub.Register("Tyvin", 67)
		stub.Register("Tyrion", 37)
		Ω(stub.RegisterCallCount()).Should(Equal(2))
	})

	It("is possible to stub the behavior", func() {
		stub.RegisterStub = func(name string, age int) {
			methodWasCalled = true
			methodFirstArg = name
			methodSecondArg = age
		}
		stub.Register("Jon", 20)
		Ω(methodWasCalled).Should(BeTrue())
		Ω(methodFirstArg).Should(Equal("Jon"))
		Ω(methodSecondArg).Should(Equal(20))
	})

	It("is possible to get arguments for call", func() {
		stub.Register("Tyvin", 67)
		stub.Register("Tyrion", 37)
		argName, argAge := stub.RegisterArgsForCall(0)
		Ω(argName).Should(Equal("Tyvin"))
		Ω(argAge).Should(Equal(67))

		argName, argAge = stub.RegisterArgsForCall(1)
		Ω(argName).Should(Equal("Tyrion"))
		Ω(argAge).Should(Equal(37))
	})
})
