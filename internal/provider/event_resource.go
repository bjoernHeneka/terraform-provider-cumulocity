package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &eventResource{}
var _ resource.ResourceWithImportState = &eventResource{}

type eventResource struct {
	client *client.Client
}

func NewEventResource() resource.Resource {
	return &eventResource{}
}

type eventResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SourceID     types.String `tfsdk:"source_id"`
	Type         types.String `tfsdk:"type"`
	Text         types.String `tfsdk:"text"`
	Time         types.String `tfsdk:"time"`
	CreationTime types.String `tfsdk:"creation_time"`
	LastUpdated  types.String `tfsdk:"last_updated"`
}

func (r *eventResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event"
}

func (r *eventResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity event. Events represent time-stamped occurrences on a device or asset. " +
			"Only the text description can be updated after creation; changing source, type, or time forces a new resource. " +
			"Corresponds to POST/GET/PUT/DELETE /event/events/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: "The managed object (device/asset) ID to which the event is associated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Event type identifier, e.g. `c8y_LocationUpdate`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable description of the event.",
			},
			"time": schema.StringAttribute{
				Required:    true,
				Description: "ISO 8601 date-time of when the event occurred, e.g. `2024-01-15T10:30:00.000Z`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the event was created in Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of the last update.",
			},
		},
	}
}

func (r *eventResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *eventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := r.client.CreateEvent(ctx,
		plan.SourceID.ValueString(),
		plan.Type.ValueString(),
		plan.Text.ValueString(),
		plan.Time.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating event", err.Error())
		return
	}

	r.apiToState(ev, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *eventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state eventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := r.client.GetEvent(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading event", err.Error())
		return
	}

	r.apiToState(ev, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *eventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan eventResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ev, err := r.client.UpdateEvent(ctx, plan.ID.ValueString(), plan.Text.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating event", err.Error())
		return
	}

	r.apiToState(ev, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *eventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEvent(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting event", err.Error())
	}
}

func (r *eventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ev, err := r.client.GetEvent(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing event", err.Error())
		return
	}
	var state eventResourceModel
	r.apiToState(ev, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *eventResource) apiToState(e *client.Event, m *eventResourceModel) {
	m.ID = types.StringValue(e.ID)
	m.SourceID = types.StringValue(e.Source.ID)
	m.Type = types.StringValue(e.Type)
	m.Text = types.StringValue(e.Text)
	m.Time = types.StringValue(e.Time)
	m.CreationTime = types.StringValue(e.CreationTime)
	m.LastUpdated = types.StringValue(e.LastUpdated)
}
