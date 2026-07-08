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

var _ resource.Resource = &auditRecordResource{}
var _ resource.ResourceWithImportState = &auditRecordResource{}

type auditRecordResource struct {
	client *client.Client
}

func NewAuditRecordResource() resource.Resource {
	return &auditRecordResource{}
}

type auditRecordResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SourceID     types.String `tfsdk:"source_id"`
	Activity     types.String `tfsdk:"activity"`
	Text         types.String `tfsdk:"text"`
	Time         types.String `tfsdk:"time"`
	Type         types.String `tfsdk:"type"`
	User         types.String `tfsdk:"user"`
	Application  types.String `tfsdk:"application"`
	Severity     types.String `tfsdk:"severity"`
	CreationTime types.String `tfsdk:"creation_time"`
}

func (r *auditRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_record"
}

func (r *auditRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a Cumulocity audit record. Audit records are immutable and cannot be updated or deleted — " +
			"any change to an input field will create a new audit record. " +
			"Corresponds to POST /audit/auditRecords.",
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
				Description: "The platform component or managed object ID to which the audit is associated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"activity": schema.StringAttribute{
				Required:    true,
				Description: "Summary of the action that was carried out, e.g. `Operation created`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				Required:    true,
				Description: "Detailed description of the action.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"time": schema.StringAttribute{
				Required:    true,
				Description: "ISO 8601 date-time of the audit event, e.g. `2024-01-15T10:30:00.000Z`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Platform component type. One of: Alarm, Application, BulkOperation, CepModule, Connector, Event, Group, Inventory, InventoryRole, Operation, Option, Report, SingleSignOn, SmartRule, SYSTEM, Tenant, TenantAuthConfig, TrustedCertificates, User, UserAuthentication.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Alarm", "Application", "BulkOperation", "CepModule", "Connector",
						"Event", "Group", "Inventory", "InventoryRole", "Operation",
						"Option", "Report", "SingleSignOn", "SmartRule", "SYSTEM",
						"Tenant", "TenantAuthConfig", "TrustedCertificates", "User", "UserAuthentication",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Optional:    true,
				Description: "The username of the user who carried out the activity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the application that performed the action (set by Cumulocity).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"severity": schema.StringAttribute{
				Computed:    true,
				Description: "Severity of the audit action (set by Cumulocity): CRITICAL, MAJOR, MINOR, WARNING, or INFORMATION.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the audit record was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *auditRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *auditRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan auditRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.CreateAuditRecord(ctx,
		plan.SourceID.ValueString(),
		plan.Activity.ValueString(),
		plan.Text.ValueString(),
		plan.Time.ValueString(),
		plan.Type.ValueString(),
		plan.User.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating audit record", err.Error())
		return
	}

	r.apiToState(record, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *auditRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state auditRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	record, err := r.client.GetAuditRecord(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading audit record", err.Error())
		return
	}

	r.apiToState(record, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is never called because all input fields have RequiresReplace.
func (r *auditRecordResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete is a no-op: Cumulocity audit records are permanent and cannot be deleted.
func (r *auditRecordResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *auditRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	record, err := r.client.GetAuditRecord(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing audit record", err.Error())
		return
	}
	var state auditRecordResourceModel
	r.apiToState(record, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *auditRecordResource) apiToState(a *client.AuditRecord, m *auditRecordResourceModel) {
	m.ID = types.StringValue(a.ID)
	m.SourceID = types.StringValue(a.Source.ID)
	m.Activity = types.StringValue(a.Activity)
	m.Text = types.StringValue(a.Text)
	m.Time = types.StringValue(a.Time)
	m.Type = types.StringValue(a.Type)
	m.User = types.StringValue(a.User)
	m.Application = types.StringValue(a.Application)
	m.Severity = types.StringValue(a.Severity)
	m.CreationTime = types.StringValue(a.CreationTime)
}
