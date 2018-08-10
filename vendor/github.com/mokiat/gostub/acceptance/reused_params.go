package acceptance

//go:generate gostub ReusedParams

type ReusedParams interface {
	Concat(first, second string)
}
