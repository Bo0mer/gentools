package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub ArraySupport

type ArraySupport interface {
	Method([3]external.Address) [3]external.Address
}
