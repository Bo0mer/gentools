// The purpose of the `external_dup` package is to duplicate (partially) the
// types and package name of the `external` folder and assure that tests fail
// if type resolution does not work correctly in gostub.
// There could be a case when the `go import` functionality that is used on
// save fixes generated stubs that were otherwise invalid. One should not
// depend on `go import` to always be able to pick the correct import path.

package external

type Address struct {
	Different bool
	Value     int
}

type Runner interface {
	Run(Address) error
}
