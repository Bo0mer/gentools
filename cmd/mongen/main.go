package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/types"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/Bo0mer/gentools/pkg/gen"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
)

const usage = `Usage: mongen [flags] <package> <interface>
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
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(2)
	}

	pkgpath, ifacename := flag.Arg(0), flag.Arg(1)

	concname := fmt.Sprintf("monitoring%s", ifacename)
	recv, err := buildReceiver(pkgpath, ifacename, concname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mogen: %s", err)
		os.Exit(1)
	}

	code := new(bytes.Buffer)
	writePackageName(code, recv)
	writeImports(code)
	recv.Interface = removePackageName(recv.Interface)
	writeConstructor(code, recv)
	writeDecl(code, recv)
	writeMethods(code, recv)

	var out = os.Stdout
	if output != "" {
		out, err = os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mogen: error creating output file: %v", err)
			os.Exit(1)
		}
		defer out.Close()
	} else {
		output = "monmw.go"
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

func writePackageName(w io.Writer, recv *gen.Receiver) {
	pkg := strings.Split(recv.Interface, ".")[0]
	fmt.Fprintf(w, "package %s\n\n", pkg)
}

func writeImports(w io.Writer) {
	fmt.Fprintf(w, `
import (
	"time"

	"github.com/go-kit/kit/metrics"
)
`)
}

func buildReceiver(pkgpath, ifacename, concname string) (*gen.Receiver, error) {
	var conf loader.Config
	conf.Import(pkgpath)
	lprog, err := conf.Load()
	if err != nil {
		return nil, err
	}
	pkg := lprog.Package(pkgpath).Pkg

	recv := &gen.Receiver{Name: "m", TypeName: concname}
	iface := pkg.Scope().Lookup(ifacename)
	if iface == nil {
		return nil, fmt.Errorf("could not find decl of %s", ifacename)
	}
	if !types.IsInterface(iface.Type()) {
		return nil, fmt.Errorf("%s is not an interface type", ifacename)
	}

	recv.Interface = pkg.Name() + "." + iface.Name()
	recv.InterfacePath = pkg.Path() + "." + iface.Name()
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

func writeConstructor(w io.Writer, recv *gen.Receiver) {

	ifaceName := recv.Interface
	if strings.Contains(ifaceName, ".") {
		ifaceName = strings.Split(ifaceName, ".")[1]
	}
	fmt.Fprintf(w, `
	// NewMonitoring%s emits metrics for executed operations. The number of
	// total operations is accumulated in totalOps, while the number of failed
	// operations is accumulated in failedOps. In addition, the duration for each
	// operation (no matter whether it failed or not) is recorded in opsDuration.
	// All measurements are labeled by operation name, thus the metrics should have
	// a single label field 'operation'.
	func NewMonitoring%s(next %s, totalOps, failedOps metrics.Counter, opsDuration metrics.Histogram) %s {
		return &%s{
			totalOps:    totalOps,
			failedOps:   failedOps,
			opsDuration: opsDuration,
			next:        next,
		}
	}`, ifaceName, ifaceName, ifaceName, ifaceName, recv.TypeName)
}

func writeDecl(w io.Writer, recv *gen.Receiver) {
	fmt.Fprintf(w, `
	// Generated using github.com/Bo0mer/gentools/cmd/mongen.
	type %s struct {
		totalOps metrics.Counter
		failedOps metrics.Counter
		opsDuration metrics.Histogram
		next %s
	}`, recv.TypeName, recv.Interface)
	fmt.Fprintln(w)
}

func writeMethods(w io.Writer, r *gen.Receiver) {
	for _, method := range r.Methods {
		writeSignature(w, r, &method)
		// method opening bracket
		fmt.Fprintln(w, "{")
		writeMethodBody(w, r, &method)

		writeReturnStatement(w, r, &method)

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
	fmt.Fprint(w, method.Results.String())
}

func writeReturnStatement(w io.Writer, r *gen.Receiver, method *gen.Method) {
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

func writeMethodBody(w io.Writer, r *gen.Receiver, method *gen.Method) {
	// Capture start time.
	fmt.Fprintln(w, `start := time.Now()`)

	// Invoke method and capture returned results.
	nRet := len(method.Results)
	for i, ret := range method.Results {
		fmt.Fprint(w, ret.Name)
		switch i {
		case nRet - 1: // last
			w.Write([]byte("="))
		default:
			w.Write([]byte(","))
		}
	}
	fmt.Fprintf(w, "%s.next.%s%s\n", r.Name, method.Name, method.Args.InvocationString())

	// Collect measurements.
	operationName := toSnakeCase(method.Name)
	fmt.Fprintf(w, `
	m.totalOps.With("operation", "%s").Add(1)
	m.opsDuration.With("operation", "%s").Observe(time.Since(start).Seconds())
	`, operationName, operationName)

	// Iff the last argument is an error, collect failure measurements if needed.
	if len(method.Results) > 0 && method.Results[nRet-1].Type == "error" {
		fmt.Fprintf(w, `
		if %s != nil {
			m.failedOps.With("operation", "%s").Add(1)
		}
		`, method.Results[nRet-1].Name, operationName)
	}

}

func toSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
