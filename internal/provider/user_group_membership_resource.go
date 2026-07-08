package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &userGroupMembershipResource{}
var _ resource.ResourceWithImportState = &userGroupMembershipResource{}

type userGroupMembershipResource struct {
	client *client.Client
}

func NewUserGroupMembershipResource() resource.Resource {
	return &userGroupMembershipResource{}
}

type userGroupMembershipModel struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	GroupID  types.Int64  `tfsdk:"group_id"`
	UserID   types.String `tfsdk:"user_id"`
}

func (r *userGroupMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group_membership"
}

func (r *userGroupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns a user to a Cumulocity user group. All attributes are immutable — any change forces recreation. " +
			"Corresponds to POST /user/{tenantId}/groups/{groupId}/users and DELETE /user/{tenantId}/groups/{groupId}/users/{userId}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {tenantId}/{groupId}/{userId}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cumulocity tenant ID. Defaults to the provider's tenant_id. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.Int64Attribute{
				Required:    true,
				Description: "Numeric ID of the user group. Use cumulocity_user_group.my_group.group_id. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "Username of the user to add to the group. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userGroupMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userGroupMembershipResource) resolveTenantID(m userGroupMembershipModel) string {
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() {
		return m.TenantID.ValueString()
	}
	return r.client.TenantID
}

func (r *userGroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userGroupMembershipModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)
	groupID := plan.GroupID.ValueInt64()
	userID := plan.UserID.ValueString()

	if err := r.client.AddUserToGroup(ctx, tenantID, groupID, userID); err != nil {
		resp.Diagnostics.AddError("Error adding user to group", err.Error())
		return
	}

	plan.TenantID = types.StringValue(tenantID)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%d/%s", tenantID, groupID, userID))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userGroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userGroupMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	exists, err := r.client.HasUserInGroup(ctx, tenantID, state.GroupID.ValueInt64(), state.UserID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		// Group itself is gone.
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading group membership", err.Error())
		return
	}
	if !exists {
		// User was removed from group outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not needed — all attributes are RequiresReplace.
func (r *userGroupMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *userGroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userGroupMembershipModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	if err := r.client.RemoveUserFromGroup(ctx, tenantID, state.GroupID.ValueInt64(), state.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing user from group", err.Error())
	}
}

// ImportState supports "{tenantId}/{groupId}/{userId}" or "{groupId}/{userId}".
func (r *userGroupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)

	var tenantID, rawGroupID, userID string
	switch len(parts) {
	case 3:
		tenantID, rawGroupID, userID = parts[0], parts[1], parts[2]
	case 2:
		tenantID = r.client.TenantID
		rawGroupID, userID = parts[0], parts[1]
	default:
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{tenantId}/{groupId}/{userId}' or '{groupId}/{userId}'.")
		return
	}

	groupID, err := strconv.ParseInt(rawGroupID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid group ID", fmt.Sprintf("group_id must be numeric, got %q", rawGroupID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), groupID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), userID)...)
}
