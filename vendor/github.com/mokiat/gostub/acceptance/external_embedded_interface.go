package acceptance

import (
	other "github.com/mokiat/gostub/acceptance/external"
	"github.com/mokiat/gostub/acceptance/external/external_dup"
)

//go:generate gostub ExternalEmbeddedInterfaceSupport

type ExternalEmbeddedInterfaceSupport interface {
	external.Runner
	Method(other.Runner) other.Runner
}
