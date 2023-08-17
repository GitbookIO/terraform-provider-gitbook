package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEntityDataSource() datasource.DataSource {
	return &entityDataSource{}
}

// entityDataSource defines the data source implementation.
type entityDataSource struct{}

func (d *entityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (d *entityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Entity data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"entity_id": schema.StringAttribute{
				Required: true,
			},
			"properties": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"string_props": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"number_props": schema.MapAttribute{
						Optional:    true,
						ElementType: types.NumberType,
					},
					"boolean_props": schema.MapAttribute{
						Optional:    true,
						ElementType: types.BoolType,
					},
				},
			},
			"urls": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"location": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (d *entityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	_, ok := req.ProviderData.(*gitbook.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gitbook.APIClient, got: %T. Please report this issue to GitBook.", req.ProviderData),
		)

		return
	}

	// TODO: Set `r.client` with Entities API client once go-gitbook is updated.
}

func (d *entityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model entityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// entityID := model.ID.ValueString()

	// entity, _, err := d.client.GetEntityById(ctx, entityID).Execute()
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error reading GitBook entity",
	// 		fmt.Sprintf("Could not fetch GitBook entity (id: %q): %v", entityID, err),
	// 	)
	// 	return
	// }

	// state.parseEntity(entity, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
