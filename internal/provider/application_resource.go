package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &applicationResource{}
var _ resource.ResourceWithImportState = &applicationResource{}

type applicationResource struct {
	client *client.Client
}

func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

type applicationModel struct {
	ID              types.String `tfsdk:"id"`
	Key             types.String `tfsdk:"key"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	ContextPath     types.String `tfsdk:"context_path"`
	Availability    types.String `tfsdk:"availability"`
	Description     types.String `tfsdk:"description"`
	Self            types.String `tfsdk:"self"`
	OwnerTenantID   types.String `tfsdk:"owner_tenant_id"`
	ActiveVersionID types.String `tfsdk:"active_version_id"`
}

func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity application. " +
			"Supported types: HOSTED (web app served by Cumulocity), EXTERNAL (link to external URL), MICROSERVICE. " +
			"Use cumulocity_application_binary to upload the ZIP archive for HOSTED/MICROSERVICE applications. " +
			"Corresponds to POST/GET/PUT/DELETE /application/applications/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the application assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Unique application key used as an identifier across tenants, e.g. \"my-app-key\".",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the application.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Application type: HOSTED, EXTERNAL, or MICROSERVICE. Immutable — changing forces a new resource.",
				Validators: []validator.String{
					stringvalidator.OneOf("HOSTED", "EXTERNAL", "MICROSERVICE"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"context_path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL context path that makes the application accessible, e.g. \"myapp\". Required by Cumulocity for HOSTED applications.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Access level for other tenants: MARKET (visible in marketplace) or PRIVATE (owner tenant only). Defaults to PRIVATE.",
				Validators: []validator.String{
					stringvalidator.OneOf("MARKET", "PRIVATE"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable description of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the tenant that owns this application.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active_version_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the currently active binary version. Set after a successful cumulocity_application_binary upload.",
			},
		},
	}
}

func (r *applicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.CreateApplication(
		ctx,
		plan.Key.ValueString(),
		plan.Name.ValueString(),
		plan.Type.ValueString(),
		plan.ContextPath.ValueString(),
		plan.Availability.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application", err.Error())
		return
	}

	r.apiToState(app, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	r.apiToState(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan applicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state applicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	app, err := r.client.UpdateApplication(
		ctx,
		plan.ID.ValueString(),
		plan.Key.ValueString(),
		plan.Name.ValueString(),
		plan.ContextPath.ValueString(),
		plan.Availability.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application", err.Error())
		return
	}

	r.apiToState(app, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteApplication(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting application", err.Error())
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *applicationResource) apiToState(app *client.Application, m *applicationModel) {
	m.ID = types.StringValue(app.ID)
	m.Key = types.StringValue(app.Key)
	m.Name = types.StringValue(app.Name)
	m.Type = types.StringValue(app.Type)
	m.ContextPath = types.StringValue(app.ContextPath)
	m.Availability = types.StringValue(app.Availability)
	m.Description = types.StringValue(app.Description)
	m.Self = types.StringValue(app.Self)
	m.ActiveVersionID = types.StringValue(app.ActiveVersionID)

	ownerTenantID := ""
	if app.Owner != nil {
		ownerTenantID = app.Owner.Tenant.ID
	}
	m.OwnerTenantID = types.StringValue(ownerTenantID)
}
