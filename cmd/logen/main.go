package main

import (
	"fmt"
	"go/types"
	"io"
	"log"
	"os"

	"github.com/Bo0mer/gentools/pkg/gen"
	"golang.org/x/tools/go/loader"
)

const usage = "Usage: logen <package> <interface>"

func main() {
	if len(os.Args) != 3 {
		log.Fatal(usage)
	}
	pkgpath, ifacename := os.Args[1], os.Args[2]
	concname := fmt.Sprintf("logging%s", ifacename)

	recv, err := buildReceiver(pkgpath, ifacename, concname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logen: %s", err)
		os.Exit(1)
	}

	writeDecl(os.Stdout, recv)
	writeMethods(os.Stdout, recv)
}

func buildReceiver(pkgpath, ifacename, concname string) (*gen.Receiver, error) {
	// The loader loads a complete Go program from source code.
	var conf loader.Config
	conf.Import(pkgpath)
	lprog, err := conf.Load()
	if err != nil {
		log.Fatal(err) // load error
	}
	pkg := lprog.Package(pkgpath).Pkg

	recv := &gen.Receiver{Name: "l", TypeName: concname}
	iface := pkg.Scope().Lookup(ifacename)
	if iface == nil {
		return nil, fmt.Errorf("could not find decl of %s", ifacename)
	}
	if !types.IsInterface(iface.Type()) {
		return nil, fmt.Errorf("%s is not an interface type", ifacename)
	}

	recv.Interface = pkg.Name() + "." + iface.Name()
	ifaceType := iface.Type().Underlying().(*types.Interface)

	for i := 0; i < ifaceType.NumMethods(); i++ {
		f := ifaceType.Method(i)
		m := gen.Method{Name: f.Name()}
		s := f.Type().Underlying().(*types.Signature)

		for i := 0; i < s.Params().Len(); i++ {
			p := s.Params().At(i)
			name := p.Name()
			if name == "" {
				name = fmt.Sprintf("arg%d", i)
			}
			m.Args = append(m.Args, gen.Arg{
				Name: name,
				Type: types.TypeString(p.Type(), (*types.Package).Name),
			})
		}

		for i := 0; i < s.Results().Len(); i++ {
			r := s.Results().At(i)
			name := r.Name()
			if name == "" {
				name = fmt.Sprintf("ret%d", i)
			}

			m.Results = append(m.Results, gen.Result{
				Name: name,
				Type: types.TypeString(r.Type(), (*types.Package).Name),
			})
		}
		recv.Methods = append(recv.Methods, m)
	}

	return recv, nil
}

func writeDecl(w io.Writer, recv *gen.Receiver) {
	fmt.Fprintf(w, `type %s struct {
		log logrus.FieldLogger
		next %s
	}`, recv.TypeName, recv.Interface)
	fmt.Fprintln(w)
}

func writeMethods(w io.Writer, r *gen.Receiver) {
	for _, method := range r.Methods {
		writeSignature(os.Stdout, r, &method)
		// method opening bracket
		fmt.Fprint(w, "{")
		writeLogCall(os.Stdout, &method)

		writeReturnStatement(os.Stdout, r, &method)

		// method closing bracket
		fmt.Fprint(w, "}\n\n")
	}
}

func writeSignature(w io.Writer, r *gen.Receiver, method *gen.Method) {
	// func (f *Foer) Foo
	fmt.Fprintf(w, "func (%s *%s) %s", r.Name, r.TypeName, method.Name)
	// print arguments
	fmt.Fprintf(w, method.Args.String())
	// print return values
	fmt.Fprintf(w, method.Results.NamedString())
}

func writeReturnStatement(w io.Writer, r *gen.Receiver, method *gen.Method) {
	if len(method.Results) > 0 {
		fmt.Fprintf(w, "return ")
	}
	fmt.Fprintf(w, "%s.next.%s%s\n", r.Name, method.Name, method.Args.InvocationString())
}

func writeLogCall(w io.Writer, method *gen.Method) {
	fmt.Fprintf(w, `
		start := time.Now()
		defer func() {
			l.log.WithFields(logrus.Fields{
				"method": "%s",
				"took": time.Since(start).Seconds(),`, method.Name)
	fmt.Fprintln(w)

	for _, arg := range method.Args {
		if arg.Type == "context.Context" {
			continue
		}
		fmt.Fprintf(w, `"in_%s":%s,`, arg.Name, arg.Name)
		fmt.Fprintln(w)
	}

	for i, res := range method.Results {
		if res.Type == "error" {
			fmt.Fprintf(w, `"error":%s,`, res.Name)

			// If the method returns only error, add the trailing newline.
			if i == len(method.Results)-1 {
				fmt.Fprintln(w)
			}
			continue
		}
		fmt.Fprintf(w, `"result_%s":%s,`, res.Name, res.Name)
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "})\n}()")
}
