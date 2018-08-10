package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub PointerSupport

type PointerSupport interface {
	Method(*external.Address) *external.Address
}
