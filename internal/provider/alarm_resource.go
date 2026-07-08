package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &alarmResource{}
var _ resource.ResourceWithImportState = &alarmResource{}

type alarmResource struct {
	client *client.Client
}

func NewAlarmResource() resource.Resource {
	return &alarmResource{}
}

type alarmResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SourceID            types.String `tfsdk:"source_id"`
	Type                types.String `tfsdk:"type"`
	Text                types.String `tfsdk:"text"`
	Severity            types.String `tfsdk:"severity"`
	Status              types.String `tfsdk:"status"`
	Time                types.String `tfsdk:"time"`
	Count               types.Int64  `tfsdk:"occurrence_count"`
	CreationTime        types.String `tfsdk:"creation_time"`
	LastUpdated         types.String `tfsdk:"last_updated"`
	FirstOccurrenceTime types.String `tfsdk:"first_occurrence_time"`
}

func (r *alarmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alarm"
}

func (r *alarmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity alarm. Alarms indicate an abnormal condition that requires attention. " +
			"Corresponds to POST/GET/PUT/DELETE /alarm/alarms/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the alarm (assigned by Cumulocity).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: "The managed object ID to which the alarm is associated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Alarm type identifier, e.g. `c8y_UnavailabilityAlarm`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable description of the alarm.",
			},
			"severity": schema.StringAttribute{
				Required:    true,
				Description: "Severity of the alarm: CRITICAL, MAJOR, MINOR, or WARNING.",
				Validators: []validator.String{
					stringvalidator.OneOf("CRITICAL", "MAJOR", "MINOR", "WARNING"),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Status of the alarm: ACTIVE, ACKNOWLEDGED, or CLEARED. Defaults to ACTIVE on creation.",
				Validators: []validator.String{
					stringvalidator.OneOf("ACTIVE", "ACKNOWLEDGED", "CLEARED"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time": schema.StringAttribute{
				Required:    true,
				Description: "ISO 8601 date-time of when the alarm occurred, e.g. `2024-01-15T10:30:00.000Z`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"occurrence_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of times this alarm has been triggered (incremented on deduplication).",
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the alarm was created in Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of the last update.",
			},
			"first_occurrence_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of the first occurrence (only set when count > 1).",
			},
		},
	}
}

func (r *alarmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *alarmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alarmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alarm, err := r.client.CreateAlarm(ctx,
		plan.SourceID.ValueString(),
		plan.Type.ValueString(),
		plan.Text.ValueString(),
		plan.Severity.ValueString(),
		plan.Status.ValueString(),
		plan.Time.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alarm", err.Error())
		return
	}

	r.apiToState(alarm, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alarmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alarmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alarm, err := r.client.GetAlarm(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alarm", err.Error())
		return
	}

	r.apiToState(alarm, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alarmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alarmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alarm, err := r.client.UpdateAlarm(ctx,
		plan.ID.ValueString(),
		plan.Text.ValueString(),
		plan.Status.ValueString(),
		plan.Severity.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alarm", err.Error())
		return
	}

	r.apiToState(alarm, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alarmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alarmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAlarm(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting alarm", err.Error())
	}
}

func (r *alarmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	alarm, err := r.client.GetAlarm(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing alarm", err.Error())
		return
	}
	var state alarmResourceModel
	r.apiToState(alarm, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alarmResource) apiToState(a *client.Alarm, m *alarmResourceModel) {
	m.ID = types.StringValue(a.ID)
	m.SourceID = types.StringValue(a.Source.ID)
	m.Type = types.StringValue(a.Type)
	m.Text = types.StringValue(a.Text)
	m.Severity = types.StringValue(a.Severity)
	m.Status = types.StringValue(a.Status)
	m.Time = types.StringValue(a.Time)
	m.Count = types.Int64Value(a.Count)
	m.CreationTime = types.StringValue(a.CreationTime)
	m.LastUpdated = types.StringValue(a.LastUpdated)
	m.FirstOccurrenceTime = types.StringValue(a.FirstOccurrenceTime)
}
