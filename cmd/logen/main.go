package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Bo0mer/gentools/pkg/gen"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
)

const usage = `Usage: logen [flags] <package> <interface> <level>
  -w <file> write result to (source) file instead of stdout
`

var (
	output string
)

func init() {
	flag.StringVar(&output, "w", "", "Write output to file")
}

func main() {
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(2)
	}

	pkgpath, ifacename, level := flag.Arg(0), flag.Arg(1), flag.Arg(2)
	if level != "debug" && level != "error" {
		log.Fatal(usage)
	}
	concname := fmt.Sprintf("%sLogging%s", level, ifacename)

	recv, err := buildReceiver(pkgpath, ifacename, concname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logen: %s", err)
		os.Exit(1)
	}

	code := new(bytes.Buffer)
	writePackageName(code, recv)
	writeImports(code)
	recv.Interface = removePackageName(recv.Interface)
	writeConstructor(code, recv, level)
	writeDecl(code, recv)
	writeMethods(code, recv, level)

	var out = os.Stdout
	if output != "" {
		out, err = os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mogen: error creating output file: %v", err)
			os.Exit(1)
		}
		defer out.Close()
	} else {
		output = "logmw.go"
	}

	fmted, err := imports.Process(output, code.Bytes(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mogen: error adding imports: %v", err)
		os.Exit(1)
	}

	out.Write(fmted)
}

func removePackageName(identifier string) string {
	lastDot := strings.LastIndex(identifier, ".")
	return string(identifier[lastDot+1:])
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

func writePackageName(w io.Writer, recv *gen.Receiver) {
	pkg := strings.Split(recv.Interface, ".")[0]
	fmt.Fprintf(w, "package %s\n\n", pkg)
}

func writeImports(w io.Writer) {
	fmt.Fprintf(w, `
import (
	"time"

	"github.com/go-kit/kit/log"
)
`)
}

func writeConstructor(w io.Writer, recv *gen.Receiver, level string) {

	msg := map[string]string{
		"debug": `logs all method invocations of next, including all
// parameters and return values.
//
// DO NOT USE IN PRODUCTION ENVIRONMENTS`,
		"error": "logs all non-nil errors",
	}

	ifaceName := recv.Interface
	lvl := strings.Title(level)
	fmt.Fprintf(w, `
	// New%sLogging%s %s.
	func New%sLogging%s(next %s, log log.Logger) %s {
		return &%s{
			next:        next,
			log: log,
		}
	}
	`, lvl, ifaceName, msg[level], lvl, ifaceName, ifaceName, ifaceName, recv.TypeName)
}

func writeDecl(w io.Writer, recv *gen.Receiver) {
	fmt.Fprintf(w, `type %s struct {
		log log.Logger
		next %s
	}`, recv.TypeName, recv.Interface)
	fmt.Fprintln(w)
}

func writeMethods(w io.Writer, r *gen.Receiver, level string) {
	for _, method := range r.Methods {
		writeSignature(w, r, &method)
		// method opening bracket
		fmt.Fprint(w, "{")
		writeLogCall(w, r, &method, level)

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

func writeReturnStatementDebug(w io.Writer, r *gen.Receiver, method *gen.Method) {
	if len(method.Results) > 0 {
		fmt.Fprintf(w, "return ")
	}
	fmt.Fprintf(w, "%s.next.%s%s\n", r.Name, method.Name, method.Args.InvocationString())
}

func writeReturnStatementError(w io.Writer, r *gen.Receiver, method *gen.Method) {
	if len(method.Results) == 0 {
		return
	}
	fmt.Fprint(w, "return ")
	for i, ret := range method.Results {
		fmt.Fprint(w, ret.Name)
		if i < len(method.Results)-1 {
			w.Write([]byte(","))
		}
	}
}

func writeLogCall(w io.Writer, r *gen.Receiver, method *gen.Method, level string) {
	switch level {
	case "debug":
		writeDebugCall(w, method)
		writeReturnStatementDebug(w, r, method)
	case "error":
		writeErrorCall(w, method)
		writeReturnStatementError(w, r, method)
	default:
		panic("unknown log level " + level)
	}

}

func writeDebugCall(w io.Writer, method *gen.Method) {
	fmt.Fprintf(w, `
		start := time.Now()
		defer func() {
			l.log.Log(
				"method", "%s",
				"took", time.Since(start).Seconds(),`, method.Name)
	fmt.Fprintln(w)

	for _, arg := range method.Args {
		if arg.Type == "context.Context" {
			continue
		}
		fmt.Fprintf(w, `"in_%s",%s,`, arg.Name, arg.Name)
		fmt.Fprintln(w)
	}

	for i, res := range method.Results {
		if res.Type == "error" {
			fmt.Fprintf(w, `"error",%s,`, res.Name)

			// If the method returns only error, add the trailing newline.
			if i == len(method.Results)-1 {
				fmt.Fprintln(w)
			}
			continue
		}
		fmt.Fprintf(w, `"result_%s",%s,`, res.Name, res.Name)
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, ")\n}()")
}

func writeErrorCall(w io.Writer, method *gen.Method) {
	// Invoke method and capture returned results.
	nRet := len(method.Results)
	if nRet == 0 {
		return
	}
	for i, ret := range method.Results {
		fmt.Fprint(w, ret.Name)
		switch i {
		case nRet - 1: // last
			w.Write([]byte("="))
		default:
			w.Write([]byte(","))
		}
	}
	fmt.Fprintf(w, "l.next.%s%s\n", method.Name, method.Args.InvocationString())

	// Iff the last argument is an error, log if needed.
	lastArg := method.Results[nRet-1]
	if len(method.Results) > 0 && lastArg.Type == "error" {
		fmt.Fprintf(w, `
		if %s != nil {
			l.log.Log(
				"method", "%s",
				"error", %s.Error(),
			)
		}
		`, lastArg.Name, method.Name, lastArg.Name)
	}
}
