package acceptance

//go:generate gostub PrimitiveResults

type PrimitiveResults interface {
	User() (name string, age int, height float32)
}
