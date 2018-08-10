package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub StructSupport

type StructSupport interface {
	Method(struct {
		Input external.Address
	}) struct {
		Output external.Address
	}
}
