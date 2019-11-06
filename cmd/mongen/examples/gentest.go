package examples

import "context"

//go:generate mongen . ServiceCtx
//go:generate mongen . ServiceNoCtx
//go:generate mongen . ServiceMixedCtxMultiMethod opencensus

type ServiceCtx interface {
	DoWork(context.Context, int, string) (string, error)
}

type ServiceNoCtx interface {
	DoWork(int, string) (string, error)
}

type ServiceMixedCtxMultiMethod interface {
	DoWork(int, string) (string, error)
	DoWorkCtx(context.Context, int, string) (string, error)
}
