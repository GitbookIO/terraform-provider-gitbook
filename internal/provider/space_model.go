package provider

import (
	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type spaceModel struct {
	ID           types.String      `tfsdk:"id"`
	Type         types.String      `tfsdk:"type"`
	Title        types.String      `tfsdk:"title"`
	Visibility   types.String      `tfsdk:"visibility"`
	CreatedAt    timetypes.RFC3339 `tfsdk:"created_at"`
	UpdatedAt    timetypes.RFC3339 `tfsdk:"updated_at"`
	URLs         types.Object      `tfsdk:"urls"`
	Organization types.String      `tfsdk:"organization"`
	Parent       types.String      `tfsdk:"parent"`
}

var spaceURLsAttributeTypes = map[string]attr.Type{
	"location":  types.StringType,
	"app":       types.StringType,
	"published": types.StringType,
	"public":    types.StringType,
}

// parseSpace merges space data from GitBook into a Terraform model.
func (m *spaceModel) parseSpace(space *gitbook.Space, diags *diag.Diagnostics) {
	m.ID = types.StringValue(space.Id)
	m.Type = types.StringValue(string(space.Type))
	m.Title = types.StringValue(space.Title)
	m.Visibility = types.StringValue(string(space.Visibility))
	m.CreatedAt = timetypes.NewRFC3339Value(space.CreatedAt)
	m.UpdatedAt = timetypes.NewRFC3339Value(space.UpdatedAt)

	urls, d := types.ObjectValue(spaceURLsAttributeTypes, map[string]attr.Value{
		"location":  types.StringValue(space.Urls.Location),
		"app":       types.StringValue(space.Urls.App),
		"published": types.StringPointerValue(space.Urls.Published),
		"public":    types.StringPointerValue(space.Urls.Public),
	})
	if d.HasError() {
		diags.Append(d...)
		return
	}
	m.URLs = urls

	m.Organization = types.StringPointerValue(space.Organization)
	m.Parent = types.StringPointerValue(space.Parent)
}
