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

var _ resource.Resource = &userGroupResource{}
var _ resource.ResourceWithImportState = &userGroupResource{}

type userGroupResource struct {
	client *client.Client
}

func NewUserGroupResource() resource.Resource {
	return &userGroupResource{}
}

type userGroupModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	GroupID     types.Int64  `tfsdk:"group_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Self        types.String `tfsdk:"self"`
}

func (r *userGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (r *userGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity user group within a tenant. " +
			"User groups bundle roles and aggregate users for access control. " +
			"Corresponds to POST/GET/PUT/DELETE /user/{tenantId}/groups/{groupId}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {tenantId}/{groupId}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cumulocity tenant ID. Defaults to the provider's tenant_id.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Numeric group ID assigned by Cumulocity. Use this as group_id in cumulocity_user_group_membership.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the user group within the tenant.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Free-text description of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userGroupResource) resolveTenantID(m userGroupModel) string {
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() {
		return m.TenantID.ValueString()
	}
	return r.client.TenantID
}

func (r *userGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)
	grp, err := r.client.CreateGroup(ctx, tenantID, plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating user group", err.Error())
		return
	}

	r.apiToState(grp, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	grp, err := r.client.GetGroup(ctx, tenantID, state.GroupID.ValueInt64())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading user group", err.Error())
		return
	}

	r.apiToState(grp, tenantID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state userGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.GroupID = state.GroupID

	tenantID := r.resolveTenantID(plan)
	grp, err := r.client.UpdateGroup(ctx, tenantID, plan.GroupID.ValueInt64(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating user group", err.Error())
		return
	}

	r.apiToState(grp, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	if err := r.client.DeleteGroup(ctx, tenantID, state.GroupID.ValueInt64()); err != nil {
		resp.Diagnostics.AddError("Error deleting user group", err.Error())
	}
}

// ImportState supports "{tenantId}/{groupId}" or just "{groupId}".
func (r *userGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)

	var tenantID, rawGroupID string
	switch len(parts) {
	case 2:
		tenantID, rawGroupID = parts[0], parts[1]
	case 1:
		tenantID = r.client.TenantID
		rawGroupID = parts[0]
	default:
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{tenantId}/{groupId}' or '{groupId}'.")
		return
	}

	groupID, err := strconv.ParseInt(rawGroupID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid group ID", fmt.Sprintf("group_id must be numeric, got %q", rawGroupID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), groupID)...)
}

func (r *userGroupResource) apiToState(grp *client.Group, tenantID string, m *userGroupModel) {
	m.TenantID = types.StringValue(tenantID)
	m.GroupID = types.Int64Value(grp.ID)
	m.ID = types.StringValue(fmt.Sprintf("%s/%d", tenantID, grp.ID))
	m.Name = types.StringValue(grp.Name)
	m.Description = types.StringValue(grp.Description)
	m.Self = types.StringValue(grp.Self)
}
