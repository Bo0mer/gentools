package acceptance

//go:generate gostub ReusedResults

type ReusedResults interface {
	FullName() (first, last string)
}
