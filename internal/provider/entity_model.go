package provider

import (
	"fmt"
	"math/big"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entityModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Type           types.String `tfsdk:"type"`
	EntityID       types.String `tfsdk:"entity_id"`
	Properties     types.Map    `tfsdk:"properties"`
	URLs           types.Object `tfsdk:"urls"`
}

type entityProperty struct {
	String   types.String `tfsdk:"string"`
	Number   types.Number `tfsdk:"number"`
	Boolean  types.Bool   `tfsdk:"boolean"`
	Relation types.Object `tfsdk:"relation"`
}

var entityPropertyAttributeTypes = map[string]attr.Type{
	"string":  types.StringType,
	"number":  types.NumberType,
	"boolean": types.BoolType,
	"relation": types.ObjectType{
		AttrTypes: entityRelationPropAttributeTypes,
	},
}

type entityRelationProperty struct {
	EntityID types.String `tfsdk:"entity_id"`
}

var entityRelationPropAttributeTypes = map[string]attr.Type{
	"entity_id": types.StringType,
}

var entityURLsAttributeTypes = map[string]attr.Type{
	"location": types.StringType,
}

// parseEntity merges an Entity from GitBook into a Terraform model.
func (m *entityModel) parseEntity(entity *gitbook.Entity, diags *diag.Diagnostics) {
	m.ID = types.StringValue(entity.Id)
	m.Type = types.StringValue(entity.Type)
	m.EntityID = types.StringValue(entity.EntityId)

	urls, d := types.ObjectValue(entityURLsAttributeTypes, map[string]attr.Value{
		"location": types.StringValue(entity.Urls.Location),
	})
	if d.HasError() {
		diags.Append(d...)
		return
	}
	m.URLs = urls

	propsMap := make(map[string]attr.Value)
	for propName, propValue := range entity.Properties {
		switch actual := propValue.GetActualInstance().(type) {
		case *string:
			stringValue, d := types.ObjectValue(entityPropertyAttributeTypes, map[string]attr.Value{
				"string":   types.StringPointerValue(actual),
				"number":   types.NumberNull(),
				"boolean":  types.BoolNull(),
				"relation": types.ObjectNull(entityRelationPropAttributeTypes),
			})
			if d.HasError() {
				diags.Append(d...)
			}
			propsMap[propName] = stringValue
		case *int:
			numberValue, d := types.ObjectValue(entityPropertyAttributeTypes, map[string]attr.Value{
				"number":   types.NumberValue(big.NewFloat(float64(*actual))),
				"string":   types.StringNull(),
				"boolean":  types.BoolNull(),
				"relation": types.ObjectNull(entityRelationPropAttributeTypes),
			})
			if d.HasError() {
				diags.Append(d...)
			}
			propsMap[propName] = numberValue
		case *float32:
			numberValue, d := types.ObjectValue(entityPropertyAttributeTypes, map[string]attr.Value{
				"number":   types.NumberValue(big.NewFloat(float64(*actual))),
				"string":   types.StringNull(),
				"boolean":  types.BoolNull(),
				"relation": types.ObjectNull(entityRelationPropAttributeTypes),
			})
			if d.HasError() {
				diags.Append(d...)
			}
			propsMap[propName] = numberValue
		case *bool:
			numberValue, d := types.ObjectValue(entityPropertyAttributeTypes, map[string]attr.Value{
				"boolean": types.BoolPointerValue(actual),
				"string":  types.StringNull(),
				"number":  types.NumberNull(),
			})
			if d.HasError() {
				diags.Append(d...)
			}
			propsMap[propName] = numberValue
		case *gitbook.UpsertEntityPropertiesValueOneOf:
			relation, d := types.ObjectValue(entityRelationPropAttributeTypes, map[string]attr.Value{
				"entity_id": types.StringValue(actual.EntityId),
			})
			if d.HasError() {
				diags.Append(d...)
				continue
			}
			relationValue, d := types.ObjectValue(entityPropertyAttributeTypes, map[string]attr.Value{
				"relation": relation,
				"string":   types.StringNull(),
				"number":   types.NumberNull(),
				"boolean":  types.BoolNull(),
			})
			if d.HasError() {
				diags.Append(d...)
			}
			propsMap[propName] = relationValue
		default:
			diags.AddError(
				"Unsupported property type",
				fmt.Sprintf("Property %q has unsupported type (%T)", propName, actual),
			)
			return
		}
	}
	if diags.HasError() {
		return
	}

	props, d := types.MapValue(types.ObjectType{AttrTypes: entityPropertyAttributeTypes}, propsMap)
	if d.HasError() {
		diags.Append(d...)
		return
	}
	m.Properties = props
}
