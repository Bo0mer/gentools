package acceptance

//go:generate gostub LocalRefSupport

type LocalRefSupport interface {
	Method(Customer) Customer
}
