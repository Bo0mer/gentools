package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PrimitiveResults", func() {
	const threshold = 0.0001

	var stub *acceptance_stubs.PrimitiveResultsStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.PrimitiveResultsStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(PrimitiveResults)
		Ω(assignable).Should(BeTrue())
	})

	It("is possible to stub the behavior", func() {
		stub.UserStub = func() (name string, age int, height float32) {
			return "John", 31, 1.83
		}
		name, age, height := stub.User()
		Ω(name).Should(Equal("John"))
		Ω(age).Should(Equal(31))
		Ω(height).Should(BeNumerically("~", 1.83, threshold))
	})

	It("is possible to get call count", func() {
		stub.User()
		stub.User()
		Ω(stub.UserCallCount()).Should(Equal(2))
	})

	It("is possible to stub results", func() {
		stub.UserReturns("Jack", 53, 1.69)

		name, age, height := stub.User()
		Ω(name).Should(Equal("Jack"))
		Ω(age).Should(Equal(53))
		Ω(height).Should(BeNumerically("~", 1.69, threshold))
	})
})
