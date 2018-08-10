package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmptyInterface", func() {
	var stub *acceptance_stubs.EmptyInterfaceStub
	var other *acceptance_stubs.EmptyInterfaceStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.EmptyInterfaceStub)
		other = new(acceptance_stubs.EmptyInterfaceStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(EmptyInterface)
		Ω(assignable).Should(BeTrue())
	})

	It("two instances are by default equal", func() {
		Ω(stub).Should(Equal(other))
	})

	It("is possible to make stub instances unique", func() {
		stub.StubGUID = 1
		Ω(stub).ShouldNot(Equal(other))
	})
})
