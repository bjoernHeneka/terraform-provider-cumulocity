package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

type userResource struct {
	client *client.Client
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResourceModel mirrors the Terraform schema for a Cumulocity user.
// password and sendPasswordResetEmail are writeOnly fields: they are sent to the
// API on Create but are never returned by GET. Their state value is preserved
// across reads to avoid spurious diffs.
type userResourceModel struct {
	// Composite import ID: "{tenantId}/{userName}"
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`

	// Required on create
	UserName types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`

	// Optional user fields
	FirstName   types.String `tfsdk:"first_name"`
	LastName    types.String `tfsdk:"last_name"`
	DisplayName types.String `tfsdk:"display_name"`
	Phone       types.String `tfsdk:"phone"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Newsletter  types.Bool   `tfsdk:"newsletter"`

	// Write-only auth fields (never returned by the API)
	Password               types.String `tfsdk:"password"`
	SendPasswordResetEmail types.Bool   `tfsdk:"send_password_reset_email"`

	// Computed / read-only
	Self                types.String `tfsdk:"self"`
	PasswordStrength    types.String `tfsdk:"password_strength"`
	ShouldResetPassword types.Bool   `tfsdk:"should_reset_password"`
	LastPasswordChange  types.String `tfsdk:"last_password_change"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity user. Corresponds to POST/GET/PUT/DELETE /user/{tenantId}/users/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal Terraform ID, set to \"{tenantId}/{userName}\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The Cumulocity tenant ID in which the user is created. Defaults to the provider's tenant_id. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The user's login name. Cannot contain whitespace or +$:/ characters. Immutable — changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The user's email address.",
			},
			"first_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The user's first name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The user's last name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The user's display name shown in Cumulocity UI.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"phone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The user's phone number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the user account is enabled. Disabled users cannot log in. Defaults to true.",
			},
			"newsletter": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the user is subscribed to the Cumulocity newsletter.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The user's initial password (6–32 Latin1 characters). Write-only — never returned by the API. Either password or send_password_reset_email=true must be set on create.",
			},
			"send_password_reset_email": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, Cumulocity sends a password reset email instead of setting a password directly. Required on create if password is not set.",
			},

			// Computed (read-only from API)
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "The URI of the user resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password_strength": schema.StringAttribute{
				Computed:    true,
				Description: "Password strength indicator: GREEN, YELLOW, or RED.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"should_reset_password": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the user must reset their password on next login.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"last_password_change": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of the last password change.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)

	createReq := client.CreateUserRequest{
		UserName:    plan.UserName.ValueString(),
		Email:       plan.Email.ValueString(),
		FirstName:   plan.FirstName.ValueString(),
		LastName:    plan.LastName.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Phone:       plan.Phone.ValueString(),
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		createReq.Enabled = &v
	}
	if !plan.Newsletter.IsNull() && !plan.Newsletter.IsUnknown() {
		v := plan.Newsletter.ValueBool()
		createReq.Newsletter = &v
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() && plan.Password.ValueString() != "" {
		createReq.Password = plan.Password.ValueString()
	}
	if !plan.SendPasswordResetEmail.IsNull() && !plan.SendPasswordResetEmail.IsUnknown() && plan.SendPasswordResetEmail.ValueBool() {
		v := true
		createReq.SendPasswordResetEmail = &v
	}

	user, err := r.client.CreateUser(ctx, tenantID, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	r.apiToState(user, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	userID := state.UserName.ValueString()

	user, err := r.client.GetUser(ctx, tenantID, userID)
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	// password and send_password_reset_email are write-only — the API never
	// returns them, so we preserve whatever is already in state to avoid diffs.
	r.apiToState(user, tenantID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)
	userID := plan.UserName.ValueString()

	updateReq := client.UpdateUserRequest{
		Email:       plan.Email.ValueString(),
		FirstName:   plan.FirstName.ValueString(),
		LastName:    plan.LastName.ValueString(),
		DisplayName: plan.DisplayName.ValueString(),
		Phone:       plan.Phone.ValueString(),
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		updateReq.Enabled = &v
	}
	if !plan.Newsletter.IsNull() && !plan.Newsletter.IsUnknown() {
		v := plan.Newsletter.ValueBool()
		updateReq.Newsletter = &v
	}

	user, err := r.client.UpdateUser(ctx, tenantID, userID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	r.apiToState(user, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	userID := state.UserName.ValueString()

	if err := r.client.DeleteUser(ctx, tenantID, userID); err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
	}
}

// ImportState supports importing via "{tenantId}/{userName}" or just "{userName}".
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	var tenantID, userName string
	if len(parts) == 2 {
		tenantID = parts[0]
		userName = parts[1]
	} else {
		tenantID = r.client.TenantID
		userName = parts[0]
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), userName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), tenantID+"/"+userName)...)
}

// resolveTenantID returns the tenant ID from state/plan, falling back to
// the provider-level TenantID if the resource attribute is empty.
func (r *userResource) resolveTenantID(m userResourceModel) string {
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() && m.TenantID.ValueString() != "" {
		return m.TenantID.ValueString()
	}
	return r.client.TenantID
}

// apiToState maps an API User response back into the Terraform state model.
// Note: password and send_password_reset_email are NOT overwritten here because
// the API never returns them.
func (r *userResource) apiToState(u *client.User, tenantID string, m *userResourceModel) {
	m.TenantID = types.StringValue(tenantID)
	m.UserName = types.StringValue(u.UserName)
	m.ID = types.StringValue(tenantID + "/" + u.UserName)
	m.Email = types.StringValue(u.Email)
	m.FirstName = types.StringValue(u.FirstName)
	m.LastName = types.StringValue(u.LastName)
	m.DisplayName = types.StringValue(u.DisplayName)
	m.Phone = types.StringValue(u.Phone)
	m.Self = types.StringValue(u.Self)
	m.PasswordStrength = types.StringValue(u.PasswordStrength)
	m.LastPasswordChange = types.StringValue(u.LastPasswordChange)

	if u.Enabled != nil {
		m.Enabled = types.BoolValue(*u.Enabled)
	} else {
		m.Enabled = types.BoolValue(true) // API default
	}
	if u.Newsletter != nil {
		m.Newsletter = types.BoolValue(*u.Newsletter)
	} else {
		m.Newsletter = types.BoolValue(false)
	}
	if u.ShouldResetPassword != nil {
		m.ShouldResetPassword = types.BoolValue(*u.ShouldResetPassword)
	} else {
		m.ShouldResetPassword = types.BoolValue(false)
	}
}
