package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/numbervalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewEntityResource() resource.Resource {
	return &entityResource{}
}

type entityResource struct {
	client *gitbook.OrganizationsApiService
}

func (r *entityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (r *entityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Entity resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The computed ID of the entity. Not to be confused with the `entity_id` attribute.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization that owns the entity.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of the entity schema. Must be prefixed with `terraform:`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(entitySchemaTypeRegExp, "must be prefixed with `terraform:`"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_id": schema.StringAttribute{
				Description: "The ID of the entity, unique for the related entity schema.",
				Required:    true,
			},
			"properties": schema.MapNestedAttribute{
				Required:            true,
				MarkdownDescription: "Map of properties, where each key is the property name and the value is an object with either a `string`, `number` or `boolean` property.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"string": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(path.Expressions{
									path.MatchRelative().AtParent().AtName("number"),
									path.MatchRelative().AtParent().AtName("boolean"),
									path.MatchRelative().AtParent().AtName("relation"),
								}...),
							},
						},
						"number": schema.NumberAttribute{
							Optional: true,
							Validators: []validator.Number{
								numbervalidator.ExactlyOneOf(path.Expressions{
									path.MatchRelative().AtParent().AtName("string"),
									path.MatchRelative().AtParent().AtName("boolean"),
									path.MatchRelative().AtParent().AtName("relation"),
								}...),
							},
						},
						"boolean": schema.BoolAttribute{
							Optional: true,
							Validators: []validator.Bool{
								boolvalidator.ExactlyOneOf(path.Expressions{
									path.MatchRelative().AtParent().AtName("string"),
									path.MatchRelative().AtParent().AtName("number"),
									path.MatchRelative().AtParent().AtName("relation"),
								}...),
							},
						},
						"relation": schema.SingleNestedAttribute{
							Optional: true,
							Validators: []validator.Object{
								objectvalidator.ExactlyOneOf(path.Expressions{
									path.MatchRelative().AtParent().AtName("string"),
									path.MatchRelative().AtParent().AtName("number"),
									path.MatchRelative().AtParent().AtName("boolean"),
								}...),
							},
							Attributes: map[string]schema.Attribute{
								"entity_id": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},
			"urls": schema.SingleNestedAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"location": schema.StringAttribute{
						Computed: true,
					},
				},
			},
		},
	}
}

func (r *entityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *entityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *entityModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entity := parseUpsertEntityFromModel(ctx, *model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitbook.UpsertSchemaEntitiesRequest{
		Entities: []gitbook.UpsertEntity{*entity},
	}
	organizationID := model.OrganizationID.ValueString()
	entityType := model.Type.ValueString()
	entityID := model.EntityID.ValueString()

	// Create entity via the GitBook API.
	_, err := r.client.UpsertSchemaEntities(ctx, organizationID, entityType).UpsertSchemaEntitiesRequest(opts).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error creating GitBook entity",
			fmt.Sprintf("Could not create GitBook entity: %v", errMessage),
		)
		return
	}

	// The HTTP response when creating an entity returns `204 No Content`,
	// so we need to fetch the entity to get its (computed) properties.
	created, _, err := r.client.GetEntity(ctx, organizationID, entityType, entityID).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error reading created GitBook entity",
			fmt.Sprintf("Could not read updated GitBook entity: %v", errMessage),
		)
		return
	}

	model.parseEntity(created, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	state := &entityModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := state.OrganizationID.ValueString()
	entityType := state.Type.ValueString()
	entityID := state.EntityID.ValueString()

	entity, _, err := r.client.GetEntity(ctx, organizationID, entityType, entityID).Execute()
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

func (r *entityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *entityModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entity := parseUpsertEntityFromModel(ctx, *model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitbook.UpsertSchemaEntitiesRequest{
		Entities: []gitbook.UpsertEntity{*entity},
	}
	organizationID := model.OrganizationID.ValueString()
	entityType := model.Type.ValueString()
	entityID := model.EntityID.ValueString()

	// Create entity via the GitBook API.
	_, err := r.client.UpsertSchemaEntities(ctx, organizationID, entityType).UpsertSchemaEntitiesRequest(opts).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error updating GitBook entity",
			fmt.Sprintf("Could not update GitBook entity: %v", errMessage),
		)
		return
	}

	// The HTTP response when creating an entity returns `204 No Content`,
	// so we need to fetch the entity to get its (computed) properties.
	created, _, err := r.client.GetEntity(ctx, organizationID, entityType, entityID).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error reading updated GitBook entity",
			fmt.Sprintf("Could not read updated GitBook entity: %v", errMessage),
		)
		return
	}

	model.parseEntity(created, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model entityModel

	// Read Terraform state into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID := model.EntityID.ValueString()
	organizationID := model.OrganizationID.ValueString()
	entityType := model.Type.ValueString()

	opts := gitbook.UpsertSchemaEntitiesRequest{
		Entities: []gitbook.UpsertEntity{},
		Delete: &gitbook.UpsertSchemaEntitiesRequestDelete{
			ArrayOfString: &[]string{entityID},
		},
	}

	// Delete entity via the GitBook API.
	_, err := r.client.UpsertSchemaEntities(ctx, organizationID, entityType).UpsertSchemaEntitiesRequest(opts).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error deleting GitBook entity",
			fmt.Sprintf("Could not delete GitBook entity: %v", errMessage),
		)
	}
}

func (r *entityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func parseUpsertEntityFromModel(ctx context.Context, model entityModel, diags *diag.Diagnostics) *gitbook.UpsertEntity {
	propsState := make(map[string]entityProperty, len(model.Properties.Elements()))
	diags.Append(model.Properties.ElementsAs(ctx, &propsState, false)...)
	if diags.HasError() {
		return nil
	}

	props := make(map[string]gitbook.UpsertEntityPropertiesValue, len(propsState))
	for propName, modelPropValue := range propsState {
		propValue := gitbook.UpsertEntityPropertiesValue{
			String: modelPropValue.String.ValueStringPointer(),
			Bool:   modelPropValue.Boolean.ValueBoolPointer(),
		}
		if !modelPropValue.Number.IsNull() {
			valueFloat32, _ := modelPropValue.Number.ValueBigFloat().Float32()
			propValue.Float32 = &valueFloat32
		}
		if !modelPropValue.Relation.IsNull() {
			relation := entityRelationProperty{}
			diags.Append(modelPropValue.Relation.As(ctx, &relation, basetypes.ObjectAsOptions{})...)
			if diags.HasError() {
				return nil
			}
			propValue.UpsertEntityPropertiesValueOneOf = gitbook.NewUpsertEntityPropertiesValueOneOf(relation.EntityID.ValueString())
		}
		props[propName] = propValue
	}

	return &gitbook.UpsertEntity{
		EntityId:   model.EntityID.ValueString(),
		Properties: props,
	}
}
