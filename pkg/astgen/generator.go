package astgen

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/mokiat/gostub/generator"
	"github.com/mokiat/gostub/resolution"
	"github.com/mokiat/gostub/util"
)

// MethodConfig describes a method.
type MethodConfig struct {
	// MethodName specifies the name of the method as seen in the interface it
	// came from.
	MethodName string

	// MethodParams specifies all the parameters of the method.  They should
	// have been normalized (i.e. no type reuse and no anonymous parameters)
	// and resolved (i.e. all selector expressions resolved against the
	// generated stub's new namespace)
	MethodParams []*ast.Field

	// MethodResults specifies all the results of the method.  They should have
	// been normalized (i.e. no type reuse and no anonymous results) and
	// resolved (i.e. all selector expressions resolved against the generated
	// stub's new namespace)
	MethodResults []*ast.Field
}

func (s *MethodConfig) HasParams() bool {
	return len(s.MethodParams) > 0
}

func (s *MethodConfig) HasResults() bool {
	return len(s.MethodResults) > 0
}

type ModelBuilder interface {
	// AddMethod should add implementation for the specified method.
	AddMethod(*MethodConfig) error
}

type Generator struct {
	Model    ModelBuilder
	Locator  *resolution.Locator
	Resolver *generator.Resolver
}

func (g *Generator) ProcessInterface(d resolution.TypeDiscovery) error {
	context := resolution.NewASTFileLocatorContext(d.File, d.Location)
	iFaceType, isIFace := d.Spec.Type.(*ast.InterfaceType)
	if !isIFace {
		return errors.New(fmt.Sprintf("type '%s' in '%s' is not interface!", d.Spec.Name.String(), d.Location))
	}
	for field := range util.EachFieldInFieldList(iFaceType.Methods) {
		switch t := field.Type.(type) {
		case *ast.FuncType:
			g.processMethod(context, field.Names[0].String(), t)
		case *ast.Ident:
			g.processSubInterfaceIdent(context, t)
		case *ast.SelectorExpr:
			g.processSubInterfaceSelector(context, t)
		default:
			return errors.New("Unknown statement in interface declaration.")
		}
	}
	return nil
}

func (g *Generator) processMethod(context *resolution.LocatorContext, name string, funcType *ast.FuncType) error {
	normalizedParams, err := g.getNormalizedParams(context, funcType)
	if err != nil {
		return err
	}
	normalizedResults, err := g.getNormalizedResults(context, funcType)
	if err != nil {
		return err
	}

	source := &MethodConfig{
		MethodName:    name,
		MethodParams:  normalizedParams,
		MethodResults: normalizedResults,
	}
	err = g.Model.AddMethod(source)
	if err != nil {
		return err
	}
	return nil
}

func (g *Generator) processSubInterfaceIdent(context *resolution.LocatorContext, ident *ast.Ident) error {
	discovery, err := g.Locator.FindIdentType(context, ident)
	if err != nil {
		return err
	}
	err = g.ProcessInterface(discovery)
	if err != nil {
		return err
	}
	return nil
}

func (g *Generator) processSubInterfaceSelector(context *resolution.LocatorContext, selector *ast.SelectorExpr) error {
	discovery, err := g.Locator.FindSelectorType(context, selector)
	if err != nil {
		return err
	}
	err = g.ProcessInterface(discovery)
	if err != nil {
		return err
	}
	return nil
}

func (g *Generator) getNormalizedParams(context *resolution.LocatorContext, funcType *ast.FuncType) ([]*ast.Field, error) {
	normalizedParams := []*ast.Field{}
	paramIndex := 1
	for param := range util.EachFieldInFieldList(funcType.Params) {
		count := util.FieldTypeReuseCount(param)
		for i := 0; i < count; i++ {
			fieldName := fmt.Sprintf("arg%d", paramIndex)
			fieldType, err := g.Resolver.ResolveType(context, param.Type)
			if err != nil {
				return nil, err
			}
			normalizedParam := util.CreateField(fieldName, fieldType)
			normalizedParams = append(normalizedParams, normalizedParam)
			paramIndex++
		}
	}
	return normalizedParams, nil
}

func (g *Generator) getNormalizedResults(context *resolution.LocatorContext, funcType *ast.FuncType) ([]*ast.Field, error) {
	normalizedResults := []*ast.Field{}
	resultIndex := 1
	for result := range util.EachFieldInFieldList(funcType.Results) {
		count := util.FieldTypeReuseCount(result)
		for i := 0; i < count; i++ {
			fieldName := fmt.Sprintf("result%d", resultIndex)
			fieldType, err := g.Resolver.ResolveType(context, result.Type)
			if err != nil {
				return nil, err
			}
			normalizedResult := util.CreateField(fieldName, fieldType)
			normalizedResults = append(normalizedResults, normalizedResult)
			resultIndex++
		}
	}
	return normalizedResults, nil
}
