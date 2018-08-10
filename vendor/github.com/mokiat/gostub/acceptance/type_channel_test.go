package acceptance_test

import (
	. "github.com/mokiat/gostub/acceptance"
	"github.com/mokiat/gostub/acceptance/acceptance_stubs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypeChannel", func() {
	var stub *acceptance_stubs.ChannelSupportStub

	BeforeEach(func() {
		stub = new(acceptance_stubs.ChannelSupportStub)
	})

	It("stub is assignable to interface", func() {
		_, assignable := interface{}(stub).(ChannelSupport)
		Ω(assignable).Should(BeTrue())
	})
})
