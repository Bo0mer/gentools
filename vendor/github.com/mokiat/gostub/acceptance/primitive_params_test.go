package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrimitiveParams", func() {
	const threshold = 0.0001

	var stub *acceptance_stubs.PrimitiveParamsStub
	var runWasCalled bool
	var runCountArg int
	var runLocationArg string
	var runTimeoutArg float32

	BeforeEach(func() {
		stub = new(acceptance_stubs.PrimitiveParamsStub)
		runWasCalled = false
		runCountArg = 0
		runLocationArg = ""
		runTimeoutArg = 0.0
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(PrimitiveParams)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to stub the behavior", func() {
		stub.SaveStub = func(count int, location string, timeout float32) {
			runWasCalled = true
			runCountArg = count
			runLocationArg = location
			runTimeoutArg = timeout
		}
		stub.Save(10, "/tmp", 3.14)
		Ω(runWasCalled).Should(BeTrue())
		Ω(runCountArg).Should(Equal(10))
		Ω(runLocationArg).Should(Equal("/tmp"))
		Ω(runTimeoutArg).Should(BeNumerically("~", 3.14, threshold))
	})

	It("is possible to get call count", func() {
		stub.Save(1, "/first", 3.3)
		stub.Save(2, "/second", 5.5)
		Ω(stub.SaveCallCount()).Should(Equal(2))
	})

	It("is possible to get arguments for call", func() {
		stub.Save(1, "/first", 3.3)
		stub.Save(2, "/second", 5.5)
		argCount, argLocation, argTimeout := stub.SaveArgsForCall(0)
		Ω(argCount).Should(Equal(1))
		Ω(argLocation).Should(Equal("/first"))
		Ω(argTimeout).Should(BeNumerically("~", 3.3, threshold))

		argCount, argLocation, argTimeout = stub.SaveArgsForCall(1)
		Ω(argCount).Should(Equal(2))
		Ω(argLocation).Should(Equal("/second"))
		Ω(argTimeout).Should(BeNumerically("~", 5.5, threshold))
	})
})
