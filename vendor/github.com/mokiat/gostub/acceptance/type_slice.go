package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub SliceSupport

type SliceSupport interface {
	Method([]external.Address) []external.Address
}
