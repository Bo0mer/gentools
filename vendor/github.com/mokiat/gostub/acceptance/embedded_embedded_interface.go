package acceptance

import . "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub EmbeddedEmbeddedInterfaceSupport

type EmbeddedEmbeddedInterfaceSupport interface {
	Runner
	Method(int) int
}
