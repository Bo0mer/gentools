package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NoParamsNoResults", func() {
	var stub *acceptance_stubs.NoParamsNoResultsStub
	var runWasCalled bool

	BeforeEach(func() {
		stub = new(acceptance_stubs.NoParamsNoResultsStub)
		runWasCalled = false
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(NoParamsNoResults)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to stub the behavior", func() {
		stub.RunStub = func() {
			runWasCalled = true
		}
		stub.Run()
		Ω(runWasCalled).Should(BeTrue())
	})

	It("is possible to get call count", func() {
		stub.Run()
		stub.Run()
		Ω(stub.RunCallCount()).Should(Equal(2))
	})
})
