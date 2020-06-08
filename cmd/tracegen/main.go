package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/Bo0mer/gentools/pkg/astgen"
	"github.com/Bo0mer/gentools/pkg/resolution"
	"github.com/Bo0mer/gentools/pkg/transformation"
)

func init() {
	flag.Usage = func() {
		var out io.Writer = os.Stdout

		fmt.Fprintln(out, "A tool that generates tracing wrappers for interfaces.")
		fmt.Fprintf(out, "Usage: %s [-h] SOURCE_DIR INTERFACE_NAME\n", path.Base(os.Args[0]))
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "  Arguments:")
		fmt.Fprintln(out, "    SOURCE_DIR       Path to the file containing the interface")
		fmt.Fprintln(out, "    INTERFACE_NAME   Name of the interface which will be wrapped")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "  Options:")
		fmt.Fprintln(out, "    -h               Print this text and exit")
		fmt.Fprintln(out, "")
	}
}

func parseArgs() (sourceDir, interfaceName string, err error) {
	flag.Parse()
	if flag.NArg() != 2 {
		return "", "", errors.New("too many arguments provided")
	}

	sourceDir = flag.Arg(0)
	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return "", "", fmt.Errorf("error determining absolute path to source directory: %v", err)
	}
	interfaceName = flag.Arg(1)

	return sourceDir, interfaceName, nil
}
func main() {
	sourceDir, interfaceName, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	sourcePkgPath, err := dirToImport(sourceDir)
	if err != nil {
		log.Fatalf("error resolving import path of source directory: %v", err)
	}
	targetPkg := path.Base(sourcePkgPath) + "mws"

	locator := resolution.NewLocator()

	context := resolution.NewSingleLocationContext(sourcePkgPath)
	d, err := locator.FindIdentType(context, ast.NewIdent(interfaceName))
	if err != nil {
		log.Fatal(err)
	}

	typeName := fmt.Sprintf("tracing%s", interfaceName)

	model := newModel(sourcePkgPath, interfaceName, typeName, targetPkg)
	generator := astgen.Generator{
		Model:    model,
		Locator:  locator,
		Resolver: resolution.NewResolver(model, locator),
	}

	err = generator.ProcessInterface(d)
	if err != nil {
		log.Fatal(err)
	}

	targetPkgPath := filepath.Join(sourceDir, targetPkg)
	if err := os.MkdirAll(targetPkgPath, 0777); err != nil {
		log.Fatalf("error creating target package directory: %v", err)
	}

	fd, err := os.Create(filepath.Join(targetPkgPath, filename(interfaceName)))
	if err != nil {
		log.Fatalf("error creating output source file: %v", err)
	}
	defer fd.Close()

	err = model.WriteSource(fd)
	if err != nil {
		log.Fatal(err)
	}

	wd, _ := os.Getwd()
	path, err := filepath.Rel(wd, fd.Name())
	if err != nil {
		path = fd.Name()
	}
	fmt.Printf("Wrote tracing implementation of %q to %q\n", sourcePkgPath+"."+interfaceName, path)
}

func filename(interfaceName string) string {
	return fmt.Sprintf("tracing_%s.go", transformation.ToSnakeCase(interfaceName))
}

func dirToImport(p string) (string, error) {
	pkg, err := build.ImportDir(p, build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.ImportPath, nil
}

func importToDir(imp string) (string, error) {
	pkg, err := build.Import(imp, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.Dir, nil
}
