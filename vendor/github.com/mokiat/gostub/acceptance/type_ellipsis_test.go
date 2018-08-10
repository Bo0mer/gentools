package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"
	"github.com/mokiat/gostub/acceptance/external/external_dup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypeEllipsis", func() {
	var stub *acceptance_stubs.EllipsisSupportStub
	var methodWasCalled bool
	var methodEllipsisArg []external.Address
	var firstAddress external.Address
	var secondAddress external.Address

	BeforeEach(func() {
		stub = new(acceptance_stubs.EllipsisSupportStub)
		methodWasCalled = false
		methodEllipsisArg = []external.Address{}
		firstAddress = external.Address{
			Value: 1,
		}
		secondAddress = external.Address{
			Value: 2,
		}
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(EllipsisSupport)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to stub the behavior", func() {
		stub.MethodStub = func(arg1 string, arg2 int, ellipsis ...external.Address) {
			methodWasCalled = true
			methodEllipsisArg = ellipsis
		}
		stub.Method("whatever", 0, firstAddress, secondAddress)
		Ω(methodWasCalled).Should(BeTrue())
		Ω(methodEllipsisArg).Should(Equal([]external.Address{firstAddress, secondAddress}))
	})

	It("is possible to get call count", func() {
		stub.Method("whatever", 0, firstAddress, secondAddress)
		stub.Method("whatever", 0, firstAddress, secondAddress)
		Ω(stub.MethodCallCount()).Should(Equal(2))
	})

	It("is possible to get arguments for call", func() {
		stub.Method("first", 1, firstAddress)
		stub.Method("second", 2, firstAddress, secondAddress)

		_, _, argAddresses := stub.MethodArgsForCall(0)
		Ω(argAddresses).Should(Equal([]external.Address{firstAddress}))

		_, _, argAddresses = stub.MethodArgsForCall(1)
		Ω(argAddresses).Should(Equal([]external.Address{firstAddress, secondAddress}))
	})
})
