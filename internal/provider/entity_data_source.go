package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func NewEntityDataSource() datasource.DataSource {
	return &entityDataSource{}
}

// entityDataSource defines the data source implementation.
type entityDataSource struct {
	client *gitbook.OrganizationsApiService
}

func (d *entityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (d *entityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Entity data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"entity_id": schema.StringAttribute{
				Required: true,
			},
			"properties": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string": schema.StringAttribute{
							Optional: true,
						},
						"number": schema.NumberAttribute{
							Optional: true,
						},
						"boolean": schema.BoolAttribute{
							Optional: true,
						},
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

	client, ok := req.ProviderData.(*gitbook.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gitbook.APIClient, got: %T. Please report this issue to GitBook.", req.ProviderData),
		)

		return
	}

	d.client = client.OrganizationsApi
}

func (d *entityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state.
	state := &entityModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := state.OrganizationID.ValueString()
	entityType := state.Type.ValueString()
	entityID := state.EntityID.ValueString()

	entity, _, err := d.client.GetEntity(ctx, organizationID, entityType, entityID).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error reading GitBook entity",
			fmt.Sprintf("Could not read GitBook entity: %v", errMessage),
		)
		return
	}

	state.parseEntity(entity, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
