package acceptance

type Customer struct {
	Name    string
	Address string
}

type Scheduler interface {
	Schedule(string, Customer) int
}
