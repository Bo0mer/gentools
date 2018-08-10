package acceptance

//go:generate gostub AnonymousResults

type AnonymousResults interface {
	ActiveUser() (int, string)
}
