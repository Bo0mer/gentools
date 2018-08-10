package acceptance

import . "github.com/mokiat/gostub/acceptance/embedded"

//go:generate gostub EmbeddedRefSupport

type EmbeddedRefSupport interface {
	Method(Resource) Resource
}
