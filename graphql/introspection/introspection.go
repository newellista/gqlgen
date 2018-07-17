//go:generate go run ./inliner/inliner.go

// introspection implements the spec defined in https://github.com/facebook/graphql/blob/master/spec/Section%204%20--%20Introspection.md#schema-introspection
package introspection

import "github.com/vektah/gqlparser/ast"

type (
	Directive struct {
		Name        string
		Description string
		Locations   []string
		Args        []InputValue
	}

	EnumValue struct {
		Name              string
		Description       string
		IsDeprecated      bool
		DeprecationReason string
	}

	Field struct {
		Name              string
		Description       string
		Type              *Type
		Args              []InputValue
		IsDeprecated      bool
		DeprecationReason string
	}

	InputValue struct {
		Name         string
		Description  string
		DefaultValue string
		Type         *Type
	}
)

func WrapSchema(schema *ast.Schema) *Schema {
	return &Schema{schema: schema}
}

func isDeprecated(directives ast.Directives) bool {
	return directives.Get("deprecated") != nil
}

func deprecationReason(directives ast.Directives) string {
	deprecation := directives.Get("deprecated")
	if deprecation == nil {
		return ""
	}

	reason := deprecation.GetArg("reason")
	if reason == nil {
		return ""
	}

	return reason.Value.Raw
}
