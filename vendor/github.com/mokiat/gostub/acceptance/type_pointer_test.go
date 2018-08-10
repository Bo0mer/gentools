package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypePointer", func() {
	var stub *acceptance_stubs.PointerSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.PointerSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(PointerSupport)
		Ω(assignable).Should(BeTrue())
	})
})
