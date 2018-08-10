package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MismatchedRefSupport", func() {
	var stub *acceptance_stubs.MismatchedRefSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.MismatchedRefSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(MismatchedRefSupport)
		Ω(assignable).Should(BeTrue())
	})
})
