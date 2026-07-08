package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &newDeviceRequestResource{}
var _ resource.ResourceWithImportState = &newDeviceRequestResource{}

type newDeviceRequestResource struct {
	client *client.Client
}

func NewNewDeviceRequestResource() resource.Resource {
	return &newDeviceRequestResource{}
}

type newDeviceRequestModel struct {
	ID            types.String `tfsdk:"id"`
	DeviceID      types.String `tfsdk:"device_id"`
	GroupID       types.String `tfsdk:"group_id"`
	DeviceType    types.String `tfsdk:"device_type"`
	Status        types.String `tfsdk:"status"`
	SecurityToken types.String `tfsdk:"security_token"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Owner         types.String `tfsdk:"owner"`
	CreationTime  types.String `tfsdk:"creation_time"`
	Self          types.String `tfsdk:"self"`
}

func (r *newDeviceRequestResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_new_device_request"
}

func (r *newDeviceRequestResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity device registration request. " +
			"A device connects to Cumulocity, receives WAITING_FOR_CONNECTION status, then moves to " +
			"PENDING_ACCEPTANCE once it establishes a connection. Set status = ACCEPTED to approve the device. " +
			"Corresponds to POST/GET/PUT/DELETE /devicecontrol/newDeviceRequests/{requestId}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier (equals device_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "External ID of the device to register. Used as the request identifier. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the device group the device will be assigned to upon acceptance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Device type, e.g. \"c8y_Linux\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Registration status. Set to ACCEPTED to approve the device. Values: WAITING_FOR_CONNECTION, PENDING_ACCEPTANCE, ACCEPTED.",
				Validators: []validator.String{
					stringvalidator.OneOf("WAITING_FOR_CONNECTION", "PENDING_ACCEPTANCE", "ACCEPTED"),
				},
			},
			"security_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Security token verified against the token submitted by the device. Required when accepting a device if security token policy is enforced. Write-only — never read back from the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant that owns this device registration request.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Username of the request owner.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the request was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the device registration request.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *newDeviceRequestResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *newDeviceRequestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan newDeviceRequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the desired status before apiToState overwrites it.
	// Cumulocity always creates requests in WAITING_FOR_CONNECTION — the status
	// can only transition to ACCEPTED after the device connects (PENDING_ACCEPTANCE).
	// The Update function handles the actual status transition.
	plannedStatus := plan.Status

	ndr, err := r.client.CreateNewDeviceRequest(ctx,
		plan.DeviceID.ValueString(),
		plan.GroupID.ValueString(),
		plan.DeviceType.ValueString(),
		plan.SecurityToken.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating new device request", err.Error())
		return
	}
	r.apiToState(ndr, &plan)

	// Preserve the planned status in state so Terraform does not see an
	// inconsistency. The Read will sync the real status on the next refresh,
	// and a subsequent apply will trigger Update once the device has connected.
	if !plannedStatus.IsNull() && !plannedStatus.IsUnknown() {
		plan.Status = plannedStatus
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *newDeviceRequestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state newDeviceRequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ndr, err := r.client.GetNewDeviceRequest(ctx, state.DeviceID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading new device request", err.Error())
		return
	}

	r.apiToState(ndr, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *newDeviceRequestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan newDeviceRequestModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ndr, err := r.client.UpdateNewDeviceRequest(ctx,
		plan.DeviceID.ValueString(),
		plan.Status.ValueString(),
		plan.SecurityToken.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating new device request", err.Error())
		return
	}

	r.apiToState(ndr, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *newDeviceRequestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state newDeviceRequestModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNewDeviceRequest(ctx, state.DeviceID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting new device request", err.Error())
	}
}

func (r *newDeviceRequestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), req.ID)...)
}

func (r *newDeviceRequestResource) apiToState(ndr *client.NewDeviceRequest, m *newDeviceRequestModel) {
	m.ID = types.StringValue(ndr.ID)
	m.DeviceID = types.StringValue(ndr.ID)
	m.GroupID = types.StringValue(ndr.GroupID)
	m.DeviceType = types.StringValue(ndr.Type)
	m.Status = types.StringValue(ndr.Status)
	m.TenantID = types.StringValue(ndr.TenantID)
	m.Owner = types.StringValue(ndr.Owner)
	m.CreationTime = types.StringValue(ndr.CreationTime)
	m.Self = types.StringValue(ndr.Self)
	// security_token is write-only; preserve existing state value.
}
