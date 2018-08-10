package acceptance

//go:generate gostub AnonymousParams

type AnonymousParams interface {
	Register(string, int)
}
