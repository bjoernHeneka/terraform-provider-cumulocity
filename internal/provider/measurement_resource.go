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

var _ resource.Resource = &measurementResource{}
var _ resource.ResourceWithImportState = &measurementResource{}

type measurementResource struct {
	client *client.Client
}

func NewMeasurementResource() resource.Resource {
	return &measurementResource{}
}

type measurementResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SourceID     types.String `tfsdk:"source_id"`
	Type         types.String `tfsdk:"type"`
	Time         types.String `tfsdk:"time"`
	Fragments    types.String `tfsdk:"fragments"`
	CreationTime types.String `tfsdk:"creation_time"`
	Self         types.String `tfsdk:"self"`
}

func (r *measurementResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_measurement"
}

func (r *measurementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Cumulocity measurement. Measurements are immutable — all attributes " +
			"trigger a resource replacement when changed. " +
			"Corresponds to POST/GET/DELETE /measurement/measurements/{id}.",
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
				Description: "The managed object (device/asset) ID to which the measurement belongs. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Measurement type, e.g. `c8y_TemperatureMeasurement`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"time": schema.StringAttribute{
				Required:    true,
				Description: "ISO 8601 date-time of when the measurement was taken, e.g. `2024-01-15T10:30:00.000Z`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fragments": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JSON object containing the measurement fragment data, e.g. `{\"c8y_Temperature\":{\"T\":{\"value\":22.5,\"unit\":\"°C\"}}}`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the measurement was created in Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the measurement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *measurementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *measurementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan measurementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.CreateMeasurement(ctx,
		plan.SourceID.ValueString(),
		plan.Type.ValueString(),
		plan.Time.ValueString(),
		plan.Fragments.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating measurement", err.Error())
		return
	}

	r.apiToState(m, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *measurementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state measurementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetMeasurement(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading measurement", err.Error())
		return
	}

	r.apiToState(m, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all attributes are RequiresReplace.
func (r *measurementResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *measurementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state measurementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteMeasurement(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting measurement", err.Error())
	}
}

func (r *measurementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	m, err := r.client.GetMeasurement(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing measurement", err.Error())
		return
	}
	var state measurementResourceModel
	r.apiToState(m, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *measurementResource) apiToState(m *client.Measurement, model *measurementResourceModel) {
	model.ID = types.StringValue(m.ID)
	model.SourceID = types.StringValue(m.Source.ID)
	model.Type = types.StringValue(m.Type)
	model.Time = types.StringValue(m.Time)
	model.CreationTime = types.StringValue(m.CreationTime)
	model.Self = types.StringValue(m.Self)
	model.Fragments = types.StringValue(m.FragmentsJSON)
}
