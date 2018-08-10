package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub InterfaceSupport

type InterfaceSupport interface {
	Method(interface {
		external.Runner
		ResolveAddress(external.Address) external.Address
	}) interface {
		external.Runner
		ProcessAddress(external.Address) external.Address
	}
}
