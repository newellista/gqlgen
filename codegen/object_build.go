package codegen

import (
	"log"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/ast"
	"golang.org/x/tools/go/loader"
)

func (cfg *Config) buildObjects(types NamedTypes, prog *loader.Program, imports *Imports) (Objects, error) {
	var objects Objects

	for _, typ := range cfg.schema.Types {
		if typ.Kind != ast.Object {
			continue
		}

		obj, err := cfg.buildObject(types, typ)
		if err != nil {
			return nil, err
		}

		def, err := findGoType(prog, obj.Package, obj.GoType)
		if err != nil {
			return nil, err
		}
		if def != nil {
			for _, bindErr := range bindObject(def.Type(), obj, imports) {
				log.Println(bindErr.Error())
				log.Println("  Adding resolver method")
			}
		}

		objects = append(objects, obj)
	}

	sort.Slice(objects, func(i, j int) bool {
		return strings.Compare(objects[i].GQLType, objects[j].GQLType) == -1
	})

	return objects, nil
}

var keywords = []string{
	"break",
	"default",
	"func",
	"interface",
	"select",
	"case",
	"defer",
	"go",
	"map",
	"struct",
	"chan",
	"else",
	"goto",
	"package",
	"switch",
	"const",
	"fallthrough",
	"if",
	"range",
	"type",
	"continue",
	"for",
	"import",
	"return",
	"var",
}

func sanitizeGoName(name string) string {
	for _, k := range keywords {
		if name == k {
			return name + "_"
		}
	}
	return name
}

func (cfg *Config) buildObject(types NamedTypes, typ *ast.Definition) (*Object, error) {
	obj := &Object{NamedType: types[typ.Name]}
	typeEntry, entryExists := cfg.Models[typ.Name]

	for _, i := range typ.Interfaces {
		obj.Satisfies = append(obj.Satisfies, i)
	}

	for _, field := range typ.Fields {

		var forceResolver bool
		if entryExists {
			if typeField, ok := typeEntry.Fields[field.Name]; ok {
				forceResolver = typeField.Resolver
			}
		}

		var args []FieldArgument
		for _, arg := range field.Arguments {
			newArg := FieldArgument{
				GQLName:   arg.Name,
				Type:      types.getType(arg.Type),
				Object:    obj,
				GoVarName: sanitizeGoName(arg.Name),
			}

			if !newArg.Type.IsInput && !newArg.Type.IsScalar {
				return nil, errors.Errorf("%s cannot be used as argument of %s.%s. only input and scalar types are allowed", arg.Type, obj.GQLType, field.Name)
			}

			if arg.DefaultValue != nil {
				var err error
				newArg.Default, err = arg.DefaultValue.Value(nil)
				if err != nil {
					return nil, errors.Errorf("default value for %s.%s is not valid: %s", typ.Name, field.Name, err.Error())
				}
				newArg.StripPtr()
			}
			args = append(args, newArg)
		}

		obj.Fields = append(obj.Fields, Field{
			GQLName:       field.Name,
			Type:          types.getType(field.Type),
			Args:          args,
			Object:        obj,
			ForceResolver: forceResolver,
		})
	}

	if typ == cfg.schema.Query {
		obj.Root = true
	}

	if typ == cfg.schema.Mutation {
		obj.Root = true
		obj.DisableConcurrency = true
	}

	if typ == cfg.schema.Subscription {
		obj.Root = true
	}

	return obj, nil
}
