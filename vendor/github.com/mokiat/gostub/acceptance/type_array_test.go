package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypeArray", func() {
	var stub *acceptance_stubs.ArraySupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.ArraySupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(ArraySupport)
		Ω(assignable).Should(BeTrue())
	})
})
