package acceptance

import "github.com/mokiat/gostub/acceptance/external/external_dup"

//go:generate gostub EllipsisSupport

type EllipsisSupport interface {
	Method(string, int, ...external.Address)
}
