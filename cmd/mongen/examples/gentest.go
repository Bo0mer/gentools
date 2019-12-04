package examples

import "context"

//go:generate mongen . GoKitService go-kit
//go:generate mongen . OCService opencensus

type GoKitService interface {
	DoWork(int, string) (string, error)
	DoWorkCtx(context.Context, int, string) (string, error)
}

type OCService interface {
	DoWork(int, string) (string, error)
	DoWorkCtx(context.Context, int, string) (string, error)
}
