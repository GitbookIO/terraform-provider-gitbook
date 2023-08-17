package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entityModel struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	EntityID   types.String `tfsdk:"entity_id"`
	Properties types.Object `tfsdk:"properties"`
	URLs       types.Object `tfsdk:"urls"`
}

var entityURLsAttributeTypes = map[string]attr.Type{
	"location": types.StringType,
}
