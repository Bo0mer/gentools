package acceptance

//go:generate gostub PrimitiveParams

type PrimitiveParams interface {
	Save(count int, location string, timeout float32)
}
