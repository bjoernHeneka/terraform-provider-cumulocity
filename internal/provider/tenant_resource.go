package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &tenantResource{}
var _ resource.ResourceWithImportState = &tenantResource{}

type tenantResource struct {
	client *client.Client
}

func NewTenantResource() resource.Resource {
	return &tenantResource{}
}

type tenantModel struct {
	ID                 types.String `tfsdk:"id"`
	AdminEmail         types.String `tfsdk:"admin_email"`
	AdminName          types.String `tfsdk:"admin_name"`
	AdminPass          types.String `tfsdk:"admin_pass"`
	Company            types.String `tfsdk:"company"`
	ContactName        types.String `tfsdk:"contact_name"`
	ContactPhone       types.String `tfsdk:"contact_phone"`
	Domain             types.String `tfsdk:"domain"`
	Parent             types.String `tfsdk:"parent"`
	Status             types.String `tfsdk:"status"`
	CreationTime       types.String `tfsdk:"creation_time"`
	AllowCreateTenants types.Bool   `tfsdk:"allow_create_tenants"`
	Self               types.String `tfsdk:"self"`
}

func (r *tenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *tenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity subtenant. " +
			"Corresponds to POST/GET/PUT/DELETE /tenant/tenants/{tenantId}. " +
			"Requires ROLE_TENANT_MANAGEMENT_ADMIN or ROLE_TENANT_MANAGEMENT_CREATE.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The tenant ID assigned by Cumulocity (e.g. \"t0071234\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"company": schema.StringAttribute{
				Required:    true,
				Description: "The tenant's company name.",
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The tenant's domain (e.g. \"mytenant.cumulocity.com\"). Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_email": schema.StringAttribute{
				Required:    true,
				Description: "Email address of the tenant administrator.",
			},
			"admin_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Username of the tenant administrator. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_pass": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password of the tenant administrator. Write-only — not returned by the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"contact_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the contact person.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"contact_phone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Phone number of the contact person in international format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the parent tenant.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the tenant: ACTIVE or SUSPENDED.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_time": schema.StringAttribute{
				Computed:    true,
				Description: "Date and time when the tenant was created (RFC 3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_create_tenants": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this tenant is allowed to create subtenants.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the tenant.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *tenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.TenantCreateRequest{
		AdminEmail:   plan.AdminEmail.ValueString(),
		AdminName:    plan.AdminName.ValueString(),
		AdminPass:    plan.AdminPass.ValueString(),
		Company:      plan.Company.ValueString(),
		ContactName:  plan.ContactName.ValueString(),
		ContactPhone: plan.ContactPhone.ValueString(),
		Domain:       plan.Domain.ValueString(),
	}

	t, err := r.client.CreateTenant(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating tenant", err.Error())
		return
	}

	r.apiToState(t, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.client.GetTenant(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading tenant", err.Error())
		return
	}

	r.apiToState(t, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.TenantUpdateRequest{
		AdminEmail:   plan.AdminEmail.ValueString(),
		Company:      plan.Company.ValueString(),
		ContactName:  plan.ContactName.ValueString(),
		ContactPhone: plan.ContactPhone.ValueString(),
		AdminPass:    plan.AdminPass.ValueString(),
	}

	t, err := r.client.UpdateTenant(ctx, plan.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating tenant", err.Error())
		return
	}

	r.apiToState(t, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTenant(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting tenant", err.Error())
	}
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *tenantResource) apiToState(t *client.Tenant, m *tenantModel) {
	m.ID = types.StringValue(t.ID)
	m.Self = types.StringValue(t.Self)
	m.Company = types.StringValue(t.Company)
	m.Domain = types.StringValue(t.Domain)
	m.Parent = types.StringValue(t.Parent)
	m.Status = types.StringValue(t.Status)
	m.CreationTime = types.StringValue(t.CreationTime)
	m.AllowCreateTenants = types.BoolValue(t.AllowCreateTenants)
	// The Cumulocity API does not echo admin_email, admin_name, contact_name,
	// or contact_phone in Create/Update responses. Only overwrite when the API
	// actually returns a value so the plan value is preserved after apply.
	if t.AdminEmail != "" {
		m.AdminEmail = types.StringValue(t.AdminEmail)
	}
	if t.AdminName != "" {
		m.AdminName = types.StringValue(t.AdminName)
	}
	if t.ContactName != "" {
		m.ContactName = types.StringValue(t.ContactName)
	}
	if t.ContactPhone != "" {
		m.ContactPhone = types.StringValue(t.ContactPhone)
	}
}
