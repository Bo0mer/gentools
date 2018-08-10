package acceptance

//go:generate gostub LocalEmbeddedInterfaceSupport

type LocalEmbeddedInterfaceSupport interface {
	Scheduler
	Method(int) int
}
