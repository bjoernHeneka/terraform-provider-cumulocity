package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &managedObjectResource{}
var _ resource.ResourceWithImportState = &managedObjectResource{}

type managedObjectResource struct {
	client *client.Client
}

func NewManagedObjectResource() resource.Resource {
	return &managedObjectResource{}
}

type managedObjectModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	Owner         types.String `tfsdk:"owner"`
	Self          types.String `tfsdk:"self"`
	CreationTime  types.String `tfsdk:"creation_time"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	IsDevice      types.Bool   `tfsdk:"is_device"`
	IsDeviceGroup types.Bool   `tfsdk:"is_device_group"`
}

func (r *managedObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_object"
}

func (r *managedObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity managed object (device, group, or generic asset) in the inventory. " +
			"The managed object ID can be used as managed_object_id in cumulocity_user_inventory_role_assignment. " +
			"Corresponds to POST/GET/PUT/DELETE /inventory/managedObjects/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the managed object assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the managed object.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Device class type. Devices with the same type can receive the same configuration, software, and operations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_device": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, adds the c8y_IsDevice fragment, marking this object as a device.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_device_group": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, adds the c8y_IsDeviceGroup fragment, marking this object as a device group.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Username of the managed object's owner. Set by the server from the authenticated user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the managed object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the managed object was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the managed object was last updated.",
			},
		},
	}
}

func (r *managedObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *managedObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan managedObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mo, err := r.client.CreateManagedObject(
		ctx,
		plan.Name.ValueString(),
		plan.Type.ValueString(),
		plan.IsDevice.ValueBool(),
		plan.IsDeviceGroup.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating managed object", err.Error())
		return
	}

	r.apiToState(mo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *managedObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state managedObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mo, err := r.client.GetManagedObject(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading managed object", err.Error())
		return
	}

	r.apiToState(mo, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *managedObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan managedObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state managedObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	mo, err := r.client.UpdateManagedObject(
		ctx,
		plan.ID.ValueString(),
		plan.Name.ValueString(),
		plan.Type.ValueString(),
		plan.IsDevice.ValueBool(),
		plan.IsDeviceGroup.ValueBool(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating managed object", err.Error())
		return
	}

	r.apiToState(mo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *managedObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state managedObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteManagedObject(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting managed object", err.Error())
	}
}

// ImportState supports both "{id}" and "id" (passthrough).
func (r *managedObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected the managed object numeric ID.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// apiToState maps an API ManagedObject response into the Terraform model.
func (r *managedObjectResource) apiToState(mo *client.ManagedObject, m *managedObjectModel) {
	m.ID = types.StringValue(mo.ID)
	m.Name = types.StringValue(mo.Name)
	m.Type = types.StringValue(mo.Type)
	m.Owner = types.StringValue(mo.Owner)
	m.Self = types.StringValue(mo.Self)
	m.CreationTime = types.StringValue(mo.CreationTime)
	m.LastUpdated = types.StringValue(mo.LastUpdated)
	m.IsDevice = types.BoolValue(mo.C8yIsDevice != nil)
	m.IsDeviceGroup = types.BoolValue(mo.C8yIsDeviceGroup != nil)
}
