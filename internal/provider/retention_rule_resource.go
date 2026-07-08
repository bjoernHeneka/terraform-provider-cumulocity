package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &retentionRuleResource{}
var _ resource.ResourceWithImportState = &retentionRuleResource{}

type retentionRuleResource struct {
	client *client.Client
}

func NewRetentionRuleResource() resource.Resource {
	return &retentionRuleResource{}
}

type retentionRuleModel struct {
	ID           types.String `tfsdk:"id"`
	DataType     types.String `tfsdk:"data_type"`
	FragmentType types.String `tfsdk:"fragment_type"`
	MaximumAge   types.Int64  `tfsdk:"maximum_age"`
	Source       types.String `tfsdk:"source"`
	Type         types.String `tfsdk:"type"`
	Editable     types.Bool   `tfsdk:"editable"`
	Self         types.String `tfsdk:"self"`
}

func (r *retentionRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_retention_rule"
}

func (r *retentionRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity retention rule, which controls how long data of a given type is kept. " +
			"Corresponds to POST/GET/PUT/DELETE /retention/retentions/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"maximum_age": schema.Int64Attribute{
				Required:    true,
				Description: "Maximum age of matching data, expressed in number of days.",
			},
			"data_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The data type(s) to which the rule applies. One of: `ALARM`, `AUDIT`, `BULK_OPERATION`, `EVENT`, `MEASUREMENT`, `OPERATION`, `*`. Defaults to `*`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fragment_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The fragment type(s) to which the rule applies. Used by EVENT, MEASUREMENT, OPERATION and BULK_OPERATION. Defaults to `*`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The source(s) to which the rule applies. Defaults to `*`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type(s) to which the rule applies. Used by ALARM, AUDIT, EVENT and MEASUREMENT. Defaults to `*`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"editable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the rule is editable. Set to false by the platform for system-managed rules; can only be changed by the Management tenant.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the retention rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *retentionRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *retentionRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan retentionRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.CreateRetentionRule(ctx, r.modelToRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating retention rule", err.Error())
		return
	}

	r.apiToState(rule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *retentionRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state retentionRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetRetentionRule(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading retention rule", err.Error())
		return
	}

	r.apiToState(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *retentionRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan retentionRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.UpdateRetentionRule(ctx, plan.ID.ValueString(), r.modelToRequest(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating retention rule", err.Error())
		return
	}

	r.apiToState(rule, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *retentionRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state retentionRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRetentionRule(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting retention rule", err.Error())
	}
}

func (r *retentionRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rule, err := r.client.GetRetentionRule(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing retention rule", err.Error())
		return
	}
	var state retentionRuleModel
	r.apiToState(rule, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *retentionRuleResource) modelToRequest(m retentionRuleModel) client.RetentionRuleRequest {
	return client.RetentionRuleRequest{
		DataType:     m.DataType.ValueString(),
		FragmentType: m.FragmentType.ValueString(),
		MaximumAge:   m.MaximumAge.ValueInt64(),
		Source:       m.Source.ValueString(),
		Type:         m.Type.ValueString(),
	}
}

func (r *retentionRuleResource) apiToState(rule *client.RetentionRule, m *retentionRuleModel) {
	m.ID = types.StringValue(rule.ID)
	m.DataType = types.StringValue(rule.DataType)
	m.FragmentType = types.StringValue(rule.FragmentType)
	m.MaximumAge = types.Int64Value(rule.MaximumAge)
	m.Source = types.StringValue(rule.Source)
	m.Type = types.StringValue(rule.Type)
	m.Editable = types.BoolValue(rule.Editable)
	m.Self = types.StringValue(rule.Self)
}
