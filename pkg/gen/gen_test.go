package gen_test

import (
	"testing"

	. "github.com/Bo0mer/gentools/pkg/gen"
)

func TestArgs(t *testing.T) {
	// TODO(borshukov): Switch to table tests.
	args := Args([]Arg{
		Arg{
			Name: "ctx",
			Type: "context.Context",
		},
		Arg{
			Name: "n",
			Type: "int",
		},
	})

	gotString := args.String()
	wantString := "(ctx context.Context, n int)"
	if gotString != wantString {
		t.Errorf("invalid arugments String: want %s, got: %s", wantString, gotString)
	}

	gotInvocation := args.InvocationString()
	wantInvocation := "(ctx, n)"
	if gotInvocation != wantInvocation {
		t.Errorf("invalid arugments InvocationString: want %s, got: %s", wantInvocation, gotInvocation)
	}
}

func TestResults(t *testing.T) {
	// TODO(borshukov): Switch to table tests.
	results := Results([]Result{
		Result{
			Name: "n",
			Type: "int",
		},
		Result{
			Name: "err",
			Type: "error",
		},
	})

	gotString := results.NamedString()
	wantString := "(n int, err error)"
	if gotString != wantString {
		t.Errorf("invalid results NamedString: want %s, got: %s", wantString, gotString)
	}
}
