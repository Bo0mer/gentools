package acceptance

import custom "github.com/mokiat/gostub/acceptance/aliased"

//go:generate gostub AliasedRefSupport

type AliasedRefSupport interface {
	Method(custom.User) custom.User
}
