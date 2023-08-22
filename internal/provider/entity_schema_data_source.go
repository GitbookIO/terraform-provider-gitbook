package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func NewEntitySchemaDataSource() datasource.DataSource {
	return &entitySchemaDataSource{}
}

// entitySchemaDataSource defines the data source implementation.
type entitySchemaDataSource struct {
	client *gitbook.OrganizationsApiService
}

func (d *entitySchemaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity_schema"
}

func (d *entitySchemaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Entity schema data source",

		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required: true,
			},
			"title": schema.StringAttribute{
				Required: true,
			},
			"properties": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"title": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"singular": schema.StringAttribute{
									Required: true,
								},
								"plural": schema.StringAttribute{
									Required: true,
								},
							},
						},
						"description": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},

						// Used when type is `relation`
						"entity": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"integration": schema.StringAttribute{
									Optional: true,
								},
								"type": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.ValueSetsAre(setvalidator.SizeAtLeast(1)),
				},
			},
		},
	}
}

func (d *entitySchemaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *entitySchemaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model entitySchemaModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := model.Organization.ValueString()
	entityType := model.Type.ValueString()

	// Fetch the entity schema via the GitBook API.
	entitySchema, _, err := d.client.GetEntitySchema(ctx, organizationID, entityType).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error reading GitBook entity schema",
			fmt.Sprintf("Could not fetch GitBook entity schema (organization: %q, type: %q): %v", organizationID, entityType, errMessage),
		)
		return
	}

	model.parseEntitySchema(entitySchema, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
