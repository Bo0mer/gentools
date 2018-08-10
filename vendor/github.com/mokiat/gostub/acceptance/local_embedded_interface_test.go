package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalEmbeddedInterface", func() {
	var stub *acceptance_stubs.LocalEmbeddedInterfaceSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.LocalEmbeddedInterfaceSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(LocalEmbeddedInterfaceSupport)
		Ω(assignable).Should(BeTrue())
	})
})
