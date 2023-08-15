package provider

import (
	"context"
	"fmt"

	gitbook "github.com/GitbookIO/go-gitbook/api"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func NewSpaceResource() resource.Resource {
	return &spaceResource{}
}

type spaceResource struct {
	client *gitbook.SpacesApiService
}

func (r *spaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *spaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Space resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"title": schema.StringAttribute{
				Optional: true,
			},
			"visibility": schema.StringAttribute{
				Optional: true,
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
					},
					"public": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"organization": schema.StringAttribute{
				Optional: true,
			},
			"parent": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *spaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client.SpacesApi
}

func (r *spaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *spaceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	organizationID := model.Organization.ValueString()
	if organizationID == "" {
		resp.Diagnostics.AddError("The argument \"organization\" is required to create a space.", "")
		return
	}

	spaceType := parseSpaceType(*model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitbook.RequestCreateSpace{
		Title:  model.Title.ValueStringPointer(),
		Parent: model.Parent.ValueStringPointer(),
		Type:   spaceType,
	}

	// Create space via the GitBook API.
	space, _, err := r.client.CreateSpace(ctx, organizationID).RequestCreateSpace(opts).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error creating GitBook space",
			fmt.Sprintf("Could not create GitBook space: %v", errMessage),
		)
		return
	}

	model.parseSpace(space, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *spaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	state := &spaceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceID := state.ID.ValueString()

	// Fetch the space via the GitBook API.
	space, _, err := r.client.GetSpaceById(ctx, spaceID).Execute()
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

	// Save updated state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *spaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *spaceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceType := parseSpaceType(*model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceVisibility := parseSpaceVisibility(*model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := gitbook.UpdateSpaceByIdRequest{
		Type:       spaceType,
		Visibility: spaceVisibility,
	}

	spaceID := model.ID.ValueString()

	// Update space via the GitBook API.
	space, _, err := r.client.UpdateSpaceById(ctx, spaceID).UpdateSpaceByIdRequest(opts).Execute()
	if err != nil {
		errMessage := parseErrorMessage(err)
		resp.Diagnostics.AddError(
			"Error updating GitBook space",
			fmt.Sprintf("Could not create GitBook space (id: %q): %v", spaceID, errMessage),
		)
		return
	}

	model.parseSpace(space, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *spaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Deleting a space is not supported in the GitBook API", "")
}

func parseSpaceType(model spaceModel, diags *diag.Diagnostics) *gitbook.SpaceType {
	if model.Type.ValueString() == "" {
		return nil
	}

	spaceType, err := gitbook.NewSpaceTypeFromValue(model.Type.ValueString())
	if err != nil {
		diags.AddError(
			"Invalid `type` attribute",
			fmt.Sprintf("Allowed values: %q", gitbook.AllowedSpaceTypeEnumValues),
		)
		return nil
	}
	return spaceType
}

func parseSpaceVisibility(model spaceModel, diags *diag.Diagnostics) *gitbook.ContentVisibility {
	if model.Visibility.ValueString() == "" {
		return nil
	}

	visibility, err := gitbook.NewContentVisibilityFromValue(model.Visibility.ValueString())
	if err != nil {
		diags.AddError(
			"Invalid space visibility",
			fmt.Sprintf("Allowed values: %q", gitbook.AllowedContentVisibilityEnumValues),
		)
		return nil
	}
	return visibility
}

func (r *spaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
