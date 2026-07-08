package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &bulkOperationResource{}
var _ resource.ResourceWithImportState = &bulkOperationResource{}

type bulkOperationResource struct {
	client *client.Client
}

func NewBulkOperationResource() resource.Resource {
	return &bulkOperationResource{}
}

type bulkOperationModel struct {
	ID                     types.String  `tfsdk:"id"`
	GroupID                types.String  `tfsdk:"group_id"`
	FailedParentID         types.String  `tfsdk:"failed_parent_id"`
	StartDate              types.String  `tfsdk:"start_date"`
	CreationRamp           types.Float64 `tfsdk:"creation_ramp"`
	OperationPrototypeJSON types.String  `tfsdk:"operation_prototype_json"`
	Status                 types.String  `tfsdk:"status"`
	GeneralStatus          types.String  `tfsdk:"general_status"`
	Self                   types.String  `tfsdk:"self"`
	ProgressPending        types.Int64   `tfsdk:"progress_pending"`
	ProgressFailed         types.Int64   `tfsdk:"progress_failed"`
	ProgressExecuting      types.Int64   `tfsdk:"progress_executing"`
	ProgressSuccessful     types.Int64   `tfsdk:"progress_successful"`
	ProgressAll            types.Int64   `tfsdk:"progress_all"`
}

func (r *bulkOperationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_operation"
}

func (r *bulkOperationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity bulk operation that sends an operation to all devices in a group. " +
			"group_id and failed_parent_id are mutually exclusive. " +
			"start_date, creation_ramp, and operation_prototype_json can be updated in-place. " +
			"Corresponds to POST/GET/PUT/DELETE /devicecontrol/bulkoperations/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the bulk operation assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the device group to target. Mutually exclusive with failed_parent_id. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"failed_parent_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of a previous bulk operation; reschedules only its failed operations. Mutually exclusive with group_id. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_date": schema.StringAttribute{
				Required:    true,
				Description: "ISO 8601 datetime when the individual operations should start being created, e.g. \"2025-01-01T12:00:00Z\".",
			},
			"creation_ramp": schema.Float64Attribute{
				Required:    true,
				Description: "Delay in seconds between creation of consecutive individual operations.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"operation_prototype_json": schema.StringAttribute{
				Required:    true,
				Description: "JSON object representing the operation to send to each device, e.g. \"{\\\"c8y_Restart\\\":{}}\".",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Internal execution status: ACTIVE, IN_PROGRESS, COMPLETED, or DELETED.",
			},
			"general_status": schema.StringAttribute{
				Computed:    true,
				Description: "End-user status: SCHEDULED, EXECUTING, EXECUTING_WITH_ERRORS, SUCCESSFUL, FAILED, or CANCELED.",
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the bulk operation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"progress_pending": schema.Int64Attribute{
				Computed:      true,
				Description:   "Number of pending individual operations.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"progress_failed": schema.Int64Attribute{
				Computed:      true,
				Description:   "Number of failed individual operations.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"progress_executing": schema.Int64Attribute{
				Computed:      true,
				Description:   "Number of individual operations currently executing.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"progress_successful": schema.Int64Attribute{
				Computed:      true,
				Description:   "Number of successfully completed individual operations.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"progress_all": schema.Int64Attribute{
				Computed:      true,
				Description:   "Total number of individual operations in this bulk operation.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *bulkOperationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bulkOperationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bulkOperationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bo, err := r.client.CreateBulkOperation(ctx,
		plan.GroupID.ValueString(),
		plan.FailedParentID.ValueString(),
		plan.StartDate.ValueString(),
		plan.CreationRamp.ValueFloat64(),
		plan.OperationPrototypeJSON.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bulk operation", err.Error())
		return
	}

	r.apiToState(bo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bulkOperationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bulkOperationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bo, err := r.client.GetBulkOperation(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading bulk operation", err.Error())
		return
	}

	r.apiToState(bo, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *bulkOperationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bulkOperationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state bulkOperationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	bo, err := r.client.UpdateBulkOperation(ctx,
		plan.ID.ValueString(),
		plan.StartDate.ValueString(),
		plan.CreationRamp.ValueFloat64(),
		plan.OperationPrototypeJSON.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating bulk operation", err.Error())
		return
	}

	r.apiToState(bo, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *bulkOperationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bulkOperationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBulkOperation(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting bulk operation", err.Error())
	}
}

func (r *bulkOperationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *bulkOperationResource) apiToState(bo *client.BulkOperation, m *bulkOperationModel) {
	m.ID = types.StringValue(bo.ID)
	m.GroupID = types.StringValue(bo.GroupID)
	m.FailedParentID = types.StringValue(bo.FailedParentID)
	m.StartDate = types.StringValue(bo.StartDate)
	m.CreationRamp = types.Float64Value(bo.CreationRamp)
	m.Status = types.StringValue(bo.Status)
	m.GeneralStatus = types.StringValue(bo.GeneralStatus)
	m.Self = types.StringValue(bo.Self)

	// Preserve planned operation_prototype_json — the API may not return it verbatim.
	if bo.OperationPrototype != nil && string(bo.OperationPrototype) != "{}" && string(bo.OperationPrototype) != "null" {
		m.OperationPrototypeJSON = types.StringValue(string(bo.OperationPrototype))
	}

	pending, failed, executing, successful, all := int64(0), int64(0), int64(0), int64(0), int64(0)
	if bo.Progress != nil {
		pending = int64(bo.Progress.Pending)
		failed = int64(bo.Progress.Failed)
		executing = int64(bo.Progress.Executing)
		successful = int64(bo.Progress.Successful)
		all = int64(bo.Progress.All)
	}
	m.ProgressPending = types.Int64Value(pending)
	m.ProgressFailed = types.Int64Value(failed)
	m.ProgressExecuting = types.Int64Value(executing)
	m.ProgressSuccessful = types.Int64Value(successful)
	m.ProgressAll = types.Int64Value(all)
}
