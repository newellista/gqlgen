package introspection

import (
	"github.com/vektah/gqlparser/ast"
)

type Type struct {
	schema *ast.Schema
	def    *ast.Definition
	typ    *ast.Type
}

func WrapTypeFromDef(s *ast.Schema, def *ast.Definition) *Type {
	if def == nil {
		return nil
	}
	return &Type{schema: s, def: def}
}

func WrapTypeFromType(s *ast.Schema, typ *ast.Type) *Type {
	if typ == nil {
		return nil
	}

	if !typ.NonNull && typ.NamedType != "" {
		return &Type{schema: s, def: s.Types[typ.NamedType]}
	}
	return &Type{schema: s, typ: typ}
}

func (t *Type) Kind() string {
	if t.def != nil {
		return string(t.def.Kind)
	}

	if t.typ.NonNull {
		return "NOT_NULL"
	}

	if t.typ.Elem != nil {
		return "LIST"
	}
	return "UNKNOWN"
}

func (t *Type) Name() string {
	if t.def == nil {
		return ""
	}
	return t.def.Name
}

func (t *Type) Description() string {
	if t.def == nil {
		return ""
	}
	return t.def.Description
}

func (t *Type) Fields(includeDeprecated bool) []Field {
	if t.def == nil || (t.def.Kind != ast.Object && t.def.Kind != ast.Interface) {
		return nil
	}
	var fields []Field
	for _, f := range t.def.Fields {
		fields = append(fields, Field{
			Name:              f.Name,
			Description:       f.Description,
			Type:              WrapTypeFromType(t.schema, f.Type),
			IsDeprecated:      isDeprecated(f.Directives),
			DeprecationReason: deprecationReason(f.Directives),
		})
	}
	return fields
}

func (t *Type) InputFields() []InputValue {
	if t.def == nil || t.def.Kind != ast.InputObject {
		return nil
	}

	var res []InputValue
	for _, f := range t.def.Fields {
		res = append(res, InputValue{
			Name:        f.Name,
			Description: f.Description,
			Type:        WrapTypeFromType(t.schema, f.Type),
		})
	}
	return res
}

func (t *Type) Interfaces() []Type {
	if t.def == nil || t.def.Kind != ast.Object {
		return nil
	}

	var res []Type
	for _, intf := range t.def.Interfaces {
		res = append(res, *WrapTypeFromDef(t.schema, t.schema.Types[intf]))
	}

	return res
}

func (t *Type) PossibleTypes() []Type {
	if t.def == nil || (t.def.Kind != ast.Interface && t.def.Kind != ast.Union) {
		return nil
	}

	var res []Type
	for _, pt := range t.schema.GetPossibleTypes(t.def) {
		res = append(res, *WrapTypeFromDef(t.schema, pt))
	}
	return res
}

func (t *Type) EnumValues(includeDeprecated bool) []EnumValue {
	if t.def == nil || t.def.Kind != ast.Enum {
		return nil
	}

	var res []EnumValue
	for _, val := range t.def.Values {
		res = append(res, EnumValue{
			Name:              val.Name,
			Description:       val.Description,
			IsDeprecated:      isDeprecated(val.Directives),
			DeprecationReason: deprecationReason(val.Directives),
		})
	}
	return res
}

func (t *Type) OfType() *Type {
	if t.typ.NonNull {
		// fake non null nodes
		cpy := t.typ
		cpy.NonNull = false

		return WrapTypeFromType(t.schema, cpy)
	}
	return WrapTypeFromType(t.schema, t.typ.Elem)
}
