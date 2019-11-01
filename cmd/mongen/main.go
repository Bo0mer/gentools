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
)

const (
	goKitProvider      = "go-kit"
	opensensusProvider = "opensensus"
)

type args struct {
	sourceDir          string
	interfaceName      string
	monitoringProvider string
}

func isValidProvider(provider string) bool {
	return provider == goKitProvider || provider == opensensusProvider
}

func parseArgs() (args, error) {
	flag.Parse()
	if flag.NArg() < 2 {
		return args{}, errors.New("too few arguments provided")
	}

	sourceDir := flag.Arg(0)
	sourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return args{}, fmt.Errorf("error determining absolute path to source directory: %v", err)
	}
	interfaceName := flag.Arg(1)

	monitoringProvider := goKitProvider
	if flag.NArg() == 3 {
		monitoringProvider = flag.Arg(2)
		if !isValidProvider(monitoringProvider) {
			return args{}, fmt.Errorf("unknown monitoring provider: %s", monitoringProvider)
		}
	}

	return args{
		sourceDir:          sourceDir,
		interfaceName:      interfaceName,
		monitoringProvider: monitoringProvider,
	}, nil
}

func main() {
	args, err := parseArgs()
	if err != nil {
		log.Fatal(err)
	}

	sourcePkgPath, err := dirToImport(args.sourceDir)
	if err != nil {
		log.Fatalf("error resolving import path of source directory: %v", err)
	}
	targetPkg := path.Base(sourcePkgPath) + "mws"

	locator := resolution.NewLocator()

	context := resolution.NewSingleLocationContext(sourcePkgPath)
	d, err := locator.FindIdentType(context, ast.NewIdent(args.interfaceName))
	if err != nil {
		log.Fatal(err)
	}

	typeName := fmt.Sprintf("monitoring%s", args.interfaceName)

	model, err := newModel(args.monitoringProvider, sourcePkgPath, args.interfaceName, typeName, targetPkg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	generator := astgen.Generator{
		Model:    model,
		Locator:  locator,
		Resolver: resolution.NewResolver(model, locator),
	}

	err = generator.ProcessInterface(d)
	if err != nil {
		log.Fatal(err)
	}

	targetPkgPath := filepath.Join(args.sourceDir, targetPkg)
	if err := os.MkdirAll(targetPkgPath, 0777); err != nil {
		log.Fatalf("error creating target package directory: %v", err)
	}

	fd, err := os.Create(filepath.Join(targetPkgPath, filename(args.interfaceName)))
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
	fmt.Printf("Wrote monitoring implementation of %q to %q\n", sourcePkgPath+"."+args.interfaceName, path)
}

func filename(interfaceName string) string {
	return fmt.Sprintf("monitoring_%s.go", toSnakeCase(interfaceName))
}

func dirToImport(p string) (string, error) {
	pkg, err := build.ImportDir(p, build.FindOnly)
	if err != nil {
		return "", err
	}
	return pkg.ImportPath, nil
}

type sourceWriter interface {
	WriteSource(w io.Writer) error
}

type model interface {
	resolution.Importer
	astgen.ModelBuilder
	sourceWriter
}

func newModel(provider string, interfacePath, interfaceName, structName, targetPkg string) (model, error) {
	switch provider {
	case goKitProvider:
		return newGoKitModel(interfacePath, interfaceName, structName, targetPkg), nil
	case opensensusProvider:
		return newOpencensusModel(interfacePath, interfaceName, structName, targetPkg), nil
	}
	return nil, fmt.Errorf("unknown provider: %s", provider)
}
