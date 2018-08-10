package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AnonymousResults", func() {
	var stub *acceptance_stubs.AnonymousResultsStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.AnonymousResultsStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(AnonymousResults)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to get call count", func() {
		stub.ActiveUser()
		stub.ActiveUser()
		Ω(stub.ActiveUserCallCount()).Should(Equal(2))
	})

	It("is possible to stub the behavior", func() {
		stub.ActiveUserStub = func() (int, string) {
			return 2, "Elvis"
		}
		id, name := stub.ActiveUser()
		Ω(id).Should(Equal(2))
		Ω(name).Should(Equal("Elvis"))
	})

	It("is possible to stub results", func() {
		stub.ActiveUserReturns(131, "Forrest")

		id, name := stub.ActiveUser()
		Ω(id).Should(Equal(131))
		Ω(name).Should(Equal("Forrest"))
	})
})
