package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewEntitySchemaResource() resource.Resource {
	return &entitySchemaResource{}
}

type entitySchemaResource struct {
	client *gitbook.OrganizationsApiService
}

func (r *entitySchemaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity_schema"
}

func (r *entitySchemaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Entity schema resource",

		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
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
			"properties": schema.SetNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"title": schema.StringAttribute{
							Required: true,
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
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *entitySchemaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gitbook.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gitbook.APIClient, got: %T. Please report this issue to GitBook.", req.ProviderData),
		)

		return
	}

	r.client = client.OrganizationsApi
}

func (r *entitySchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *entitySchemaModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityRawSchema := entityRawSchemaFromModel(ctx, *model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create entity schema via the GitBook API.
	_, err := r.client.SetEntitySchema(ctx, model.Organization.ValueString(), model.Type.ValueString()).EntityRawSchema(*entityRawSchema).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error creating GitBook entity schema",
			fmt.Sprintf("Could not create GitBook entity schema: %v", errMessage),
		)
		return
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entitySchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	state := &entitySchemaModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := state.Organization.ValueString()
	entityType := state.Type.ValueString()

	// Fetch the entitySchema via the GitBook API.
	entitySchema, _, err := r.client.GetEntitySchema(ctx, organizationID, entityType).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error reading GitBook entity schema",
			fmt.Sprintf("Could not fetch GitBook entity schema (organization: %q, type: %q): %v", organizationID, entityType, errMessage),
		)
		return
	}

	state.parseEntitySchema(entitySchema, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *entitySchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *entitySchemaModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityRawSchema := entityRawSchemaFromModel(ctx, *model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update entity schema via the GitBook API.
	_, err := r.client.SetEntitySchema(ctx, model.Organization.ValueString(), model.Type.ValueString()).EntityRawSchema(*entityRawSchema).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error updating GitBook entity schema",
			fmt.Sprintf("Could not update GitBook entity schema: %v", errMessage),
		)
		return
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entitySchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model entitySchemaModel

	// Read Terraform state into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := model.Organization.ValueString()
	entityType := model.Type.ValueString()

	_, err := r.client.DeleteEntitySchema(ctx, organizationID, entityType).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error deleting GitBook entity schema",
			fmt.Sprintf("Could not delete GitBook entity schema: %v", errMessage),
		)
	}
}

func (r *entitySchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func entityRawSchemaFromModel(ctx context.Context, model entitySchemaModel, diags *diag.Diagnostics) *gitbook.EntityRawSchema {
	propsState := make([]entitySchemaProperties, 0, len(model.Properties.Elements()))
	diags.Append(model.Properties.ElementsAs(ctx, &propsState, false)...)
	if diags.HasError() {
		return nil
	}

	props := make([]gitbook.EntityPropertySchema, len(model.Properties.Elements()))
	for i, prop := range propsState {
		props[i] = gitbook.EntityPropertySchema{
			Name:        prop.Name.ValueString(),
			Title:       prop.Title.ValueString(),
			Description: prop.Description.ValueStringPointer(),
			Type:        prop.Type.ValueString(),
		}
		if !prop.Entity.IsNull() {
			entity := entitySchemaPropertyEntity{}
			diags.Append(prop.Entity.As(ctx, &entity, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil
			}
			props[i].Entity = map[string]interface{}{
				"integration": entity.Integration.ValueString(),
				"type":        entity.Type.ValueString(),
			}
		}
	}

	title := entitySchemaTitle{}
	diags.Append(model.Title.As(ctx, &title, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &gitbook.EntityRawSchema{
		Type: model.Type.ValueString(),
		Title: gitbook.EntityRawSchemaTitle{
			Singular: title.Singular.ValueString(),
			Plural:   title.Plural.ValueString(),
		},
		Properties: props,
	}
}
