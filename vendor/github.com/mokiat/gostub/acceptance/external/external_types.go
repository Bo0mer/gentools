package external

type Address struct {
	Name   string
	Number int
}

type Runner interface {
	Run()
}
