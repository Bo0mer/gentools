package gen

import (
	"bytes"
	"fmt"
)

// Receiver represents a method receiver.
type Receiver struct {
	Name          string
	TypeName      string
	Methods       []Method
	Interface     string // optional: which interface is implemented by the receiver
	InterfacePath string
}

type Method struct {
	Name    string
	Args    Args
	Results Results
}

type Args []Arg

// String returns the string representation of the function arguments.
func (a Args) String() string {
	w := bytes.NewBufferString("(") // start with opening bracket
	for i, arg := range a {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%s %s", arg.Name, arg.Type)
	}
	fmt.Fprint(w, ")") // closing bracket
	return w.String()
}

func (a Args) InvocationString() string {
	w := bytes.NewBufferString("(") // start with opening bracket
	for i, arg := range a {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%s", arg.Name)
	}
	fmt.Fprint(w, ")") // closing bracket
	return w.String()
}

type Results []Result

// String returns the string representation of the return parameters.
func (r Results) NamedString() string {
	if len(r) == 0 {
		return ""
	}
	w := bytes.NewBufferString("(") // start with opening bracket
	for i, ret := range r {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%s %s", ret.Name, ret.Type)
	}
	fmt.Fprint(w, ")") // closing bracket
	return w.String()
}

func (r Results) String() string {
	if len(r) == 0 {
		return ""
	}
	w := bytes.NewBufferString("(") // start with opening bracket
	for i, ret := range r {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprintf(w, "%s", ret.Type)
	}
	fmt.Fprint(w, ")") // closing bracket
	return w.String()
}

type Arg struct {
	Name string
	Type string
}

type Result struct {
	Name string
	Type string
}
