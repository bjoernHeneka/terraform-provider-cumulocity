package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &externalIDResource{}
var _ resource.ResourceWithImportState = &externalIDResource{}

type externalIDResource struct {
	client *client.Client
}

func NewExternalIDResource() resource.Resource {
	return &externalIDResource{}
}

type externalIDModel struct {
	ID                types.String `tfsdk:"id"`
	ExternalID        types.String `tfsdk:"external_id"`
	Type              types.String `tfsdk:"type"`
	ManagedObjectID   types.String `tfsdk:"managed_object_id"`
	ManagedObjectSelf types.String `tfsdk:"managed_object_self"`
	Self              types.String `tfsdk:"self"`
}

func (r *externalIDResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_id"
}

func (r *externalIDResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity external ID, linking a managed object to an " +
			"identifier in an external system (e.g. a device serial number). " +
			"External IDs are immutable — any change forces replacement. " +
			"Corresponds to POST /identity/globalIds/{id}/externalIds and " +
			"GET/DELETE /identity/externalIds/{type}/{externalId}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {type}/{external_id}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier value in the external system, e.g. \"SN-12345\". Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the external identifier, e.g. \"c8y_Serial\". Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_object_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the managed object this external ID is linked to. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_object_self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the linked managed object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of this external ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *externalIDResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *externalIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalIDModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ext := client.ExternalID{
		ExternalId: plan.ExternalID.ValueString(),
		Type:       plan.Type.ValueString(),
	}

	result, err := r.client.CreateExternalID(ctx, plan.ManagedObjectID.ValueString(), ext)
	if err != nil {
		resp.Diagnostics.AddError("Error creating external ID", err.Error())
		return
	}

	r.apiToState(result, plan.ManagedObjectID.ValueString(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *externalIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state externalIDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetExternalID(ctx, state.Type.ValueString(), state.ExternalID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading external ID", err.Error())
		return
	}

	moID := state.ManagedObjectID.ValueString()
	if result.ManagedObject != nil && result.ManagedObject.ID != "" {
		moID = result.ManagedObject.ID
	}
	r.apiToState(result, moID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not supported — all fields require replacement.
func (r *externalIDResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "External IDs are immutable. Destroy and recreate to change any field.")
}

func (r *externalIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state externalIDModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteExternalID(ctx, state.Type.ValueString(), state.ExternalID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting external ID", err.Error())
	}
}

// ImportState supports "{type}/{externalId}".
func (r *externalIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idx := strings.Index(req.ID, "/")
	if idx <= 0 || idx == len(req.ID)-1 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{type}/{externalId}', e.g. 'c8y_Serial/SN-12345'.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), req.ID[:idx])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("external_id"), req.ID[idx+1:])...)
}

func (r *externalIDResource) apiToState(ext *client.ExternalID, managedObjectID string, m *externalIDModel) {
	m.ExternalID = types.StringValue(ext.ExternalId)
	m.Type = types.StringValue(ext.Type)
	m.Self = types.StringValue(ext.Self)
	m.ID = types.StringValue(ext.Type + "/" + ext.ExternalId)
	m.ManagedObjectID = types.StringValue(managedObjectID)
	if ext.ManagedObject != nil {
		m.ManagedObjectSelf = types.StringValue(ext.ManagedObject.Self)
	}
}
