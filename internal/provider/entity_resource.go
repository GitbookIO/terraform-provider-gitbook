package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEntityResource() resource.Resource {
	return &entityResource{}
}

type entityResource struct{}

func (r *entityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entity"
}

func (r *entityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Entity resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_id": schema.StringAttribute{
				Required: true,
			},
			"properties": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"string_props": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
					"number_props": schema.MapAttribute{
						ElementType: types.NumberType,
						Optional:    true,
					},
					"boolean_props": schema.MapAttribute{
						ElementType: types.BoolType,
						Optional:    true,
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

	_, ok := req.ProviderData.(*gitbook.APIClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gitbook.APIClient, got: %T. Please report this issue to GitBook.", req.ProviderData),
		)

		return
	}

	// TODO: Set `r.client` with Entities API client once go-gitbook is updated.
}

func (r *entityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *entityModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// opts := gitbook.RequestCreateSpace{
	// 	Title:  model.Title.ValueStringPointer(),
	// 	Parent: model.Parent.ValueStringPointer(),
	// 	Type:   spaceType,
	// }

	// // Create space via the GitBook API.
	// space, _, err := r.client.CreateSpace(ctx, organizationID).RequestCreateSpace(opts).Execute()
	// if err != nil {
	// 	errMessage := parseErrorMessage(err)
	// 	resp.Diagnostics.AddError(
	// 		"Error creating GitBook space",
	// 		fmt.Sprintf("Could not create GitBook space: %v", errMessage),
	// 	)
	// 	return
	// }

	// model.parseSpace(space, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	state := &entityModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// entityID := state.EntityID.ValueString()

	// // Fetch the space via the GitBook API.
	// space, _, err := r.client.GetSpaceById(ctx, entityID).Execute()
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error reading GitBook space",
	// 		fmt.Sprintf("Could not fetch GitBook space (id: %q): %v", entityID, err),
	// 	)
	// 	return
	// }

	// state.parseSpace(space, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

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

	// opts := gitbook.UpdateSpaceByIdRequest{
	// 	Type:       spaceType,
	// 	Visibility: spaceVisibility,
	// }

	// spaceID := model.ID.ValueString()

	// // Update space via the GitBook API.
	// space, _, err := r.client.UpdateSpaceById(ctx, spaceID).UpdateSpaceByIdRequest(opts).Execute()
	// if err != nil {
	// 	errMessage := parseErrorMessage(err)
	// 	resp.Diagnostics.AddError(
	// 		"Error updating GitBook space",
	// 		fmt.Sprintf("Could not create GitBook space (id: %q): %v", spaceID, errMessage),
	// 	)
	// 	return
	// }

	// model.parseSpace(space, &resp.Diagnostics)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *entityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: Implement.
}

func (r *entityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
