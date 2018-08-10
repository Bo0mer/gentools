package acceptance_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

// NOTE:
// All tests in this suite should be executed through scripts/test
// which assures that the stubbing logic, that these tests verify,
// has been executed beforehand.

func TestGostub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Test Suite")
}
