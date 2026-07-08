package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &userRoleAssignmentResource{}
var _ resource.ResourceWithImportState = &userRoleAssignmentResource{}

type userRoleAssignmentResource struct {
	client *client.Client
}

func NewUserRoleAssignmentResource() resource.Resource {
	return &userRoleAssignmentResource{}
}

// userRoleAssignmentModel represents one global-role assignment on a user.
// All fields are immutable — any change forces a new resource.
type userRoleAssignmentModel struct {
	// Composite ID: "{tenantId}/{userId}/{roleId}"
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	UserID   types.String `tfsdk:"user_id"`
	Role     types.String `tfsdk:"role"`

	// Computed: self-link of the assignment returned by the API
	Self types.String `tfsdk:"self"`
}

func (r *userRoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_role_assignment"
}

func (r *userRoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns a global Cumulocity role to a user. " +
			"Corresponds to POST/DELETE /user/{tenantId}/users/{userId}/roles/{roleId}. " +
			"All fields are immutable — changing any value destroys and recreates the assignment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier: \"{tenantId}/{userId}/{role}\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The Cumulocity tenant ID. Defaults to the provider's tenant_id. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The userName of the user to assign the role to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role ID to assign, e.g. \"ROLE_ALARM_ADMIN\". Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "The self-link URL of the role assignment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userRoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userRoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userRoleAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)
	userID := plan.UserID.ValueString()
	roleID := plan.Role.ValueString()

	selfLink, err := r.client.AssignUserRole(ctx, tenantID, userID, roleID)
	if err != nil {
		resp.Diagnostics.AddError("Error assigning role to user", err.Error())
		return
	}

	plan.TenantID = types.StringValue(tenantID)
	plan.ID = types.StringValue(tenantID + "/" + userID + "/" + roleID)
	plan.Self = types.StringValue(selfLink)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userRoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	userID := state.UserID.ValueString()
	roleID := state.Role.ValueString()

	assigned, err := r.client.HasUserRole(ctx, tenantID, userID, roleID)
	if errors.Is(err, client.ErrNotFound) {
		// User itself is gone — remove this assignment from state too
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading role assignment", err.Error())
		return
	}
	if !assigned {
		resp.State.RemoveResource(ctx)
		return
	}

	// No mutable fields to refresh — state is already correct
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userRoleAssignmentResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All fields have RequiresReplace — Update is never called.
}

func (r *userRoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	if err := r.client.UnassignUserRole(ctx, tenantID, state.UserID.ValueString(), state.Role.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error unassigning role from user", err.Error())
	}
}

// ImportState supports "{tenantId}/{userId}/{roleId}" or "{userId}/{roleId}".
func (r *userRoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	var tenantID, userID, roleID string
	switch len(parts) {
	case 3:
		tenantID, userID, roleID = parts[0], parts[1], parts[2]
	case 2:
		tenantID = r.client.TenantID
		userID, roleID = parts[0], parts[1]
	default:
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected format: \"{tenantId}/{userId}/{roleId}\" or \"{userId}/{roleId}\".",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), userID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role"), roleID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), tenantID+"/"+userID+"/"+roleID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("self"), "")...)
}

func (r *userRoleAssignmentResource) resolveTenantID(m userRoleAssignmentModel) string {
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() && m.TenantID.ValueString() != "" {
		return m.TenantID.ValueString()
	}
	return r.client.TenantID
}
