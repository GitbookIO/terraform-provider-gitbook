package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewSpaceDataSource() datasource.DataSource {
	return &spaceDataSource{}
}

// spaceDataSource defines the data source implementation.
type spaceDataSource struct {
	client *gitbook.SpacesApiService
}

func (d *spaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *spaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Space data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"title": schema.StringAttribute{
				Computed: true,
			},
			"visibility": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed:   true,
				CustomType: &timetypes.RFC3339Type{},
			},
			"updated_at": schema.StringAttribute{
				Computed:   true,
				CustomType: &timetypes.RFC3339Type{},
			},
			"urls": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"location": schema.StringAttribute{
						Computed: true,
					},
					"app": schema.StringAttribute{
						Computed: true,
					},
					"published": schema.StringAttribute{
						Computed: true,
						Optional: true,
					},
					"public": schema.StringAttribute{
						Computed: true,
						Optional: true,
					},
				},
			},
			"organization": schema.StringAttribute{
				Computed: true,
			},
			"parent": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *spaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitbook.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gitbook.APIClient, got: %T. Please report this issue to GitBook.", req.ProviderData),
		)

		return
	}

	d.client = client.SpacesApi
}

func (d *spaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state spaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := state.ID.ValueString()

	space, _, err := d.client.GetSpaceById(ctx, spaceID).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading GitBook space",
			fmt.Sprintf("Could not fetch GitBook space (id: %q): %v", spaceID, err),
		)
		return
	}

	state.parseSpace(space, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
