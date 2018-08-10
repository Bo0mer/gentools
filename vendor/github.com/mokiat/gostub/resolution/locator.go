package resolution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/mokiat/gostub/util"
)

func NewLocator() *Locator {
	return &Locator{
		cache: make(map[string][]TypeDiscovery),
	}
}

type Locator struct {
	cache map[string][]TypeDiscovery
}

type TypeDiscovery struct {
	Location string
	File     *ast.File
	Spec     *ast.TypeSpec
}

func (l *Locator) FindIdentType(context *LocatorContext, ref *ast.Ident) (TypeDiscovery, error) {
	locations := context.CandidateLocations(".")
	return l.findTypeDeclarationInLocations(ref.String(), locations)
}

func (l *Locator) FindSelectorType(context *LocatorContext, ref *ast.SelectorExpr) (TypeDiscovery, error) {
	aliasIdent, ok := ref.X.(*ast.Ident)
	if !ok {
		panic("Selector expression is not a reference!")
	}
	locations := context.CandidateLocations(aliasIdent.String())
	return l.findTypeDeclarationInLocations(ref.Sel.String(), locations)
}

func (l *Locator) findTypeDeclarationInLocations(name string, candidateLocations []string) (TypeDiscovery, error) {
	for _, location := range candidateLocations {
		discovery, found, err := l.findTypeDeclarationInLocation(name, location)
		if err != nil {
			return TypeDiscovery{}, err
		}
		if found {
			return discovery, nil
		}
	}
	return TypeDiscovery{}, &TypeNotFoundError{Name: name}
}

func (l *Locator) findTypeDeclarationInLocation(name string, location string) (TypeDiscovery, bool, error) {
	discoveries, err := l.discoverTypes(location)
	if err != nil {
		return TypeDiscovery{}, false, err
	}
	for _, discovery := range discoveries {
		if discovery.Spec.Name.String() == name {
			return discovery, true, nil
		}
	}
	return TypeDiscovery{}, false, nil
}

func (l *Locator) discoverTypes(location string) ([]TypeDiscovery, error) {
	discoveries, found := l.cache[location]
	if found {
		return discoveries, nil
	}

	sourcePath, err := util.ImportToDir(location)
	if err != nil {
		return nil, err
	}

	pkgs, err := parser.ParseDir(token.NewFileSet(), sourcePath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	discoveries = make([]TypeDiscovery, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for spec := range util.EachTypeSpecificationInFile(file) {
				discoveries = append(discoveries, TypeDiscovery{
					Location: location,
					File:     file,
					Spec:     spec,
				})
			}
		}
	}
	l.cache[location] = discoveries
	return discoveries, nil
}

type TypeNotFoundError struct {
	Name string
}

func (e *TypeNotFoundError) Error() string {
	return fmt.Sprintf("Could not find '%s' type.", e.Name)
}
