package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExternalRefSupport", func() {
	var stub *acceptance_stubs.ExternalRefSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.ExternalRefSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(ExternalRefSupport)
		Ω(assignable).Should(BeTrue())
	})
})
