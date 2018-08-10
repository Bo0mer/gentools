package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmbeddedRefSupport", func() {
	var stub *acceptance_stubs.EmbeddedRefSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.EmbeddedRefSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(EmbeddedRefSupport)
		Ω(assignable).Should(BeTrue())
	})
})
