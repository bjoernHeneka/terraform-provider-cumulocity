package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &loginOptionResource{}
var _ resource.ResourceWithImportState = &loginOptionResource{}

type loginOptionResource struct {
	client *client.Client
}

func NewLoginOptionResource() resource.Resource {
	return &loginOptionResource{}
}

type loginOptionModel struct {
	ID                   types.String `tfsdk:"id"`
	Type                 types.String `tfsdk:"type"`
	ProviderName         types.String `tfsdk:"provider_name"`
	GrantType            types.String `tfsdk:"grant_type"`
	UserManagementSource types.String `tfsdk:"user_management_source"`
	VisibleOnLoginPage   types.Bool   `tfsdk:"visible_on_login_page"`
	Self                 types.String `tfsdk:"self"`

	// Computed read-back fields — reflect server state, not managed by this
	// resource. Managing full OAuth2/SSO configs is out of scope; use these to
	// inspect the effective configuration.
	Template           types.String `tfsdk:"template"`
	ButtonName         types.String `tfsdk:"button_name"`
	Issuer             types.String `tfsdk:"issuer"`
	ClientID           types.String `tfsdk:"client_id"`
	Audience           types.String `tfsdk:"audience"`
	RedirectToPlatform types.String `tfsdk:"redirect_to_platform"`
	UseIDToken         types.Bool   `tfsdk:"use_id_token"`
	ConfigJSON         types.String `tfsdk:"config_json"`
}

func (r *loginOptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login_option"
}

func (r *loginOptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity login option (authentication configuration). " +
			"Corresponds to POST/GET/PUT/DELETE /tenant/loginOptions/{typeOrId}. " +
			"Requires ROLE_TENANT_ADMIN or ROLE_TENANT_MANAGEMENT_ADMIN.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the login option assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Authentication type, e.g. BASIC, OAUTH2, OAUTH2_INTERNAL. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_name": schema.StringAttribute{
				Required:    true,
				Description: "Display name of the authentication provider shown in the UI.",
			},
			"grant_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "OAuth grant type: AUTHORIZATION_CODE or PASSWORD.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_management_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source of user management, e.g. INTERNAL or REMOTE.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visible_on_login_page": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this login option is shown on the login page.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the login option.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The configuration template, e.g. CUSTOM (OAuth2 options).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"button_name": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The label of the login button shown on the login page.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The OAuth2/OIDC token issuer URL.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The OAuth2 client ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"audience": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The OAuth2 token audience.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"redirect_to_platform": schema.StringAttribute{
				Computed:    true,
				Description: "Read-only. The platform redirect URL used in the OAuth2 flow.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_id_token": schema.BoolAttribute{
				Computed:    true,
				Description: "Read-only. Whether the ID token is used instead of the access token.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"config_json": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				Description: "Read-only. The complete raw JSON payload as returned by the API, " +
					"including all type-specific nested fields. Parse with jsondecode(). Marked " +
					"sensitive as request templates may contain secrets.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *loginOptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loginOptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loginOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := client.LoginOption{
		Type:                 plan.Type.ValueString(),
		ProviderName:         plan.ProviderName.ValueString(),
		GrantType:            plan.GrantType.ValueString(),
		UserManagementSource: plan.UserManagementSource.ValueString(),
		VisibleOnLoginPage:   plan.VisibleOnLoginPage.ValueBool(),
	}

	result, err := r.client.CreateLoginOption(ctx, opt)
	if err != nil {
		resp.Diagnostics.AddError("Error creating login option", err.Error())
		return
	}

	r.apiToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *loginOptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loginOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetLoginOption(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading login option", err.Error())
		return
	}

	r.apiToState(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *loginOptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loginOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt := client.LoginOption{
		ID:                   plan.ID.ValueString(),
		Type:                 plan.Type.ValueString(),
		ProviderName:         plan.ProviderName.ValueString(),
		GrantType:            plan.GrantType.ValueString(),
		UserManagementSource: plan.UserManagementSource.ValueString(),
		VisibleOnLoginPage:   plan.VisibleOnLoginPage.ValueBool(),
	}

	result, err := r.client.UpdateLoginOption(ctx, plan.ID.ValueString(), opt)
	if err != nil {
		resp.Diagnostics.AddError("Error updating login option", err.Error())
		return
	}

	r.apiToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *loginOptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loginOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoginOption(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting login option", err.Error())
	}
}

func (r *loginOptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *loginOptionResource) apiToState(opt *client.LoginOption, m *loginOptionModel) {
	m.ID = types.StringValue(opt.ID)
	m.Self = types.StringValue(opt.Self)
	m.Type = types.StringValue(opt.Type)
	m.ProviderName = types.StringValue(opt.ProviderName)
	m.GrantType = types.StringValue(opt.GrantType)
	m.UserManagementSource = types.StringValue(opt.UserManagementSource)
	m.VisibleOnLoginPage = types.BoolValue(opt.VisibleOnLoginPage)
	m.Template = types.StringValue(opt.Template)
	m.ButtonName = types.StringValue(opt.ButtonName)
	m.Issuer = types.StringValue(opt.Issuer)
	m.ClientID = types.StringValue(opt.ClientID)
	m.Audience = types.StringValue(opt.Audience)
	m.RedirectToPlatform = types.StringValue(opt.RedirectToPlatform)
	m.UseIDToken = types.BoolValue(opt.UseIDToken)
	m.ConfigJSON = types.StringValue(string(opt.Raw))
}
