package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReusedResults", func() {
	var stub *acceptance_stubs.ReusedResultsStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.ReusedResultsStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(ReusedResults)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to get call count", func() {
		stub.FullName()
		stub.FullName()
		Ω(stub.FullNameCallCount()).Should(Equal(2))
	})

	It("is possible to stub the behavior", func() {
		stub.FullNameStub = func() (first, last string) {
			return "Wall", "E"
		}
		first, last := stub.FullName()
		Ω(first).Should(Equal("Wall"))
		Ω(last).Should(Equal("E"))
	})

	It("is possible to stub results", func() {
		stub.FullNameReturns("Jack", "Bauer")

		first, last := stub.FullName()
		Ω(first).Should(Equal("Jack"))
		Ω(last).Should(Equal("Bauer"))
	})
})
