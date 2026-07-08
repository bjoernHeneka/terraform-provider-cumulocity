package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &deviceOperationResource{}
var _ resource.ResourceWithImportState = &deviceOperationResource{}

type deviceOperationResource struct {
	client *client.Client
}

func NewDeviceOperationResource() resource.Resource {
	return &deviceOperationResource{}
}

type deviceOperationModel struct {
	ID              types.String `tfsdk:"id"`
	DeviceID        types.String `tfsdk:"device_id"`
	Description     types.String `tfsdk:"description"`
	FragmentsJSON   types.String `tfsdk:"fragments_json"`
	Status          types.String `tfsdk:"status"`
	FailureReason   types.String `tfsdk:"failure_reason"`
	CreationTime    types.String `tfsdk:"creation_time"`
	BulkOperationID types.Int64  `tfsdk:"bulk_operation_id"`
	Self            types.String `tfsdk:"self"`
}

func (r *deviceOperationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_operation"
}

func (r *deviceOperationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sends an operation to a Cumulocity device. The operation is created in PENDING status " +
			"and progresses to EXECUTING → SUCCESSFUL or FAILED as the device processes it. " +
			"On destroy, PENDING operations are cancelled (set to FAILED); completed operations are removed from state only. " +
			"Corresponds to POST/GET /devicecontrol/operations/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the target device. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Human-readable description of the operation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fragments_json": schema.StringAttribute{
				Required: true,
				Description: "JSON object containing the operation payload, e.g. " +
					`"{\"c8y_Restart\":{}}" or "{\"c8y_Command\":{\"text\":\"ls -la\"}}". ` +
					"Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current operation status: PENDING, EXECUTING, SUCCESSFUL, or FAILED. Updated by the device.",
			},
			"failure_reason": schema.StringAttribute{
				Computed:    true,
				Description: "Reason for failure, populated when status is FAILED.",
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the operation was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bulk_operation_id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the parent bulk operation, if this operation was created by one.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the operation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *deviceOperationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceOperationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceOperationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op, err := r.client.CreateOperation(ctx,
		plan.DeviceID.ValueString(),
		plan.Description.ValueString(),
		plan.FragmentsJSON.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating device operation", err.Error())
		return
	}

	r.apiToState(op, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deviceOperationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceOperationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op, err := r.client.GetOperation(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading device operation", err.Error())
		return
	}

	r.apiToState(op, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all content attributes are RequiresReplace.
func (r *deviceOperationResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *deviceOperationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceOperationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CancelOperation(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error cancelling device operation", err.Error())
	}
}

func (r *deviceOperationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *deviceOperationResource) apiToState(op *client.Operation, m *deviceOperationModel) {
	m.ID = types.StringValue(op.ID)
	m.DeviceID = types.StringValue(op.DeviceID)
	m.Status = types.StringValue(op.Status)
	m.FailureReason = types.StringValue(op.FailureReason)
	m.CreationTime = types.StringValue(op.CreationTime)
	m.BulkOperationID = types.Int64Value(op.BulkOperationID)
	m.Self = types.StringValue(op.Self)
	// FragmentsJSON and Description are write-only — preserve from state.
}
