package provider

import (
	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entitySchemaModel struct {
	Type           types.String `tfsdk:"type"`
	Title          types.Object `tfsdk:"title"`
	Properties     types.Set    `tfsdk:"properties"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

type entitySchemaProperties struct {
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Entity      types.Object `tfsdk:"entity"`
}

var entitySchemaPropertiesAttributeTypes = map[string]attr.Type{
	"name":        types.StringType,
	"title":       types.StringType,
	"description": types.StringType,
	"type":        types.StringType,
	"entity": types.ObjectType{
		AttrTypes: entitySchemaEntityPropAttributeTypes,
	},
}

type entitySchemaTitle struct {
	Singular types.String `tfsdk:"singular"`
	Plural   types.String `tfsdk:"plural"`
}

var entitySchemaTitleAttributeTypes = map[string]attr.Type{
	"singular": types.StringType,
	"plural":   types.StringType,
}

type entitySchemaPropertyEntity struct {
	Type types.String `tfsdk:"type"`
}

var entitySchemaEntityPropAttributeTypes = map[string]attr.Type{
	"type": types.StringType,
}

// parseEntitySchema merges an entity schema from GitBook into a Terraform model.
func (m *entitySchemaModel) parseEntitySchema(entitySchema *gitbook.EntitySchema, diags *diag.Diagnostics) {
	m.Type = types.StringValue(entitySchema.Type)

	title, d := types.ObjectValue(entitySchemaTitleAttributeTypes, map[string]attr.Value{
		"singular": types.StringValue(entitySchema.Title.Singular),
		"plural":   types.StringValue(entitySchema.Title.Plural),
	})
	if d.HasError() {
		diags.Append(d...)
		return
	}
	m.Title = title

	properties := make([]attr.Value, len(entitySchema.Properties))

	for i, property := range entitySchema.Properties {
		modelPropAttributes := map[string]attr.Value{
			"name":        types.StringValue(property.Name),
			"title":       types.StringValue(property.Title),
			"description": types.StringPointerValue(property.Description),
			"type":        types.StringValue(property.Type),
		}
		if property.Entity != nil {
			entityType, _ := property.Entity["type"].(string)
			entity, d := types.ObjectValue(entitySchemaEntityPropAttributeTypes, map[string]attr.Value{
				"type": types.StringValue(entityType),
			})
			if d.HasError() {
				diags.Append(d...)
				continue
			}
			modelPropAttributes["entity"] = entity
		} else {
			modelPropAttributes["entity"] = types.ObjectNull(entitySchemaEntityPropAttributeTypes)
		}
		modelProp, d := types.ObjectValue(entitySchemaPropertiesAttributeTypes, modelPropAttributes)
		if d.HasError() {
			diags.Append(d...)
			continue
		}
		properties[i] = modelProp
	}
	if diags.HasError() {
		return
	}

	propsSetValue, d := types.SetValue(types.ObjectType{
		AttrTypes: entitySchemaPropertiesAttributeTypes,
	}, properties)
	if d.HasError() {
		diags.Append(d...)
		return
	}

	m.Properties = propsSetValue
}
