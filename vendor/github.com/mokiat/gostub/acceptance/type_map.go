package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub MapSupport

type MapSupport interface {
	Method(map[external.Address]external.Address) map[external.Address]external.Address
}
