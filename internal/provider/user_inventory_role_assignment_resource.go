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

var _ resource.Resource = &userInventoryRoleAssignmentResource{}
var _ resource.ResourceWithImportState = &userInventoryRoleAssignmentResource{}

type userInventoryRoleAssignmentResource struct {
	client *client.Client
}

func NewUserInventoryRoleAssignmentResource() resource.Resource {
	return &userInventoryRoleAssignmentResource{}
}

// userInventoryRoleAssignmentModel is the Terraform state for one inventory assignment.
// One assignment covers one user + one managed object + one or more inventory roles.
type userInventoryRoleAssignmentModel struct {
	// Composite ID: "{tenantId}/{userId}/{assignmentId}"
	ID              types.String `tfsdk:"id"`
	TenantID        types.String `tfsdk:"tenant_id"`
	UserID          types.String `tfsdk:"user_id"`
	ManagedObjectID types.String `tfsdk:"managed_object_id"`

	// The inventory role names to assign. Mutable — changing this issues a PUT.
	RoleNames types.List `tfsdk:"role_names"`

	// Computed
	AssignmentID types.Int64  `tfsdk:"assignment_id"`
	Self         types.String `tfsdk:"self"`
}

func (r *userInventoryRoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_inventory_role_assignment"
}

func (r *userInventoryRoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Assigns one or more Cumulocity inventory roles to a user for a specific managed object. " +
			"Changing role_names issues a PUT update in-place. " +
			"Changing user_id, managed_object_id or tenant_id forces a new resource. " +
			"Corresponds to POST/GET/PUT/DELETE /user/{tenantId}/users/{userId}/roles/inventory/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier: \"{tenantId}/{userId}/{assignmentId}\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Cumulocity tenant ID. Defaults to the provider's tenant_id. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The userName of the user. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_object_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the managed object (device or group) for which the roles apply. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_names": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of inventory role names to assign, e.g. [\"Operations: Restart Device\"]. Changing this list issues a PUT update.",
			},
			"assignment_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Numeric ID of the inventory assignment returned by the API.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the inventory assignment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userInventoryRoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userInventoryRoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userInventoryRoleAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(plan)
	roleNames := r.extractRoleNames(plan.RoleNames)

	assignment, err := r.client.CreateUserInventoryRoleAssignment(
		ctx, tenantID, plan.UserID.ValueString(), plan.ManagedObjectID.ValueString(), roleNames,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating inventory role assignment", err.Error())
		return
	}

	r.apiToState(assignment, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userInventoryRoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userInventoryRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	assignment, err := r.client.GetUserInventoryRoleAssignment(
		ctx, tenantID, state.UserID.ValueString(), state.AssignmentID.ValueInt64(),
	)
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading inventory role assignment", err.Error())
		return
	}

	r.apiToState(assignment, tenantID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userInventoryRoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userInventoryRoleAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the assignment_id from current state — it's stable after create.
	var state userInventoryRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AssignmentID = state.AssignmentID

	tenantID := r.resolveTenantID(plan)
	roleNames := r.extractRoleNames(plan.RoleNames)

	// PUT requires role IDs — look up each name.
	roleIDs, err := r.resolveRoleIDs(ctx, roleNames)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving inventory role IDs", err.Error())
		return
	}

	assignment, err := r.client.UpdateUserInventoryRoleAssignment(
		ctx, tenantID, plan.UserID.ValueString(), plan.AssignmentID.ValueInt64(), roleIDs,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating inventory role assignment", err.Error())
		return
	}

	r.apiToState(assignment, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userInventoryRoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userInventoryRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.resolveTenantID(state)
	if err := r.client.DeleteUserInventoryRoleAssignment(
		ctx, tenantID, state.UserID.ValueString(), state.AssignmentID.ValueInt64(),
	); err != nil {
		resp.Diagnostics.AddError("Error deleting inventory role assignment", err.Error())
	}
}

// ImportState supports "{tenantId}/{userId}/{assignmentId}" or "{userId}/{assignmentId}".
func (r *userInventoryRoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	var tenantID, userID string
	var assignmentID int64

	var rawID string
	switch len(parts) {
	case 3:
		tenantID = parts[0]
		userID = parts[1]
		rawID = parts[2]
	case 2:
		tenantID = r.client.TenantID
		userID = parts[0]
		rawID = parts[1]
	default:
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected \"{tenantId}/{userId}/{assignmentId}\" or \"{userId}/{assignmentId}\".",
		)
		return
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid assignment ID", fmt.Sprintf("%q is not a valid integer: %s", rawID, err))
		return
	}
	assignmentID = id

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), userID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("assignment_id"), assignmentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%s/%s/%d", tenantID, userID, assignmentID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("self"), "")...)
}

// resolveTenantID returns the tenant ID from model or falls back to the provider's TenantID.
func (r *userInventoryRoleAssignmentResource) resolveTenantID(m userInventoryRoleAssignmentModel) string {
	if !m.TenantID.IsNull() && !m.TenantID.IsUnknown() && m.TenantID.ValueString() != "" {
		return m.TenantID.ValueString()
	}
	return r.client.TenantID
}

// extractRoleNames converts a types.List of strings into a []string.
func (r *userInventoryRoleAssignmentResource) extractRoleNames(list types.List) []string {
	var names []string
	for _, v := range list.Elements() {
		if s, ok := v.(types.String); ok {
			names = append(names, s.ValueString())
		}
	}
	return names
}

// resolveRoleIDs looks up inventory role IDs by name for the PUT update payload.
func (r *userInventoryRoleAssignmentResource) resolveRoleIDs(ctx context.Context, roleNames []string) ([]int64, error) {
	allRoles, err := r.client.ListInventoryRoles(ctx)
	if err != nil {
		return nil, err
	}

	// Build a name→id index.
	index := make(map[string]int64, len(allRoles))
	for _, role := range allRoles {
		index[role.Name] = role.ID
	}

	ids := make([]int64, 0, len(roleNames))
	for _, name := range roleNames {
		id, ok := index[name]
		if !ok {
			return nil, fmt.Errorf("inventory role %q not found", name)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// apiToState maps an InventoryAssignment API response back into the Terraform model.
// role_names is refreshed from the API response to stay in sync.
func (r *userInventoryRoleAssignmentResource) apiToState(a *client.InventoryAssignment, tenantID string, m *userInventoryRoleAssignmentModel) {
	m.TenantID = types.StringValue(tenantID)
	m.AssignmentID = types.Int64Value(a.ID)
	m.ManagedObjectID = types.StringValue(a.ManagedObject)
	m.Self = types.StringValue(a.Self)
	m.ID = types.StringValue(fmt.Sprintf("%s/%s/%d", tenantID, m.UserID.ValueString(), a.ID))

	// Refresh role_names from the API response.
	names := make([]types.String, len(a.Roles))
	for i, role := range a.Roles {
		names[i] = types.StringValue(role.Name)
	}
	elems := make([]interface{}, len(names))
	for i, n := range names {
		elems[i] = n
	}
	listVal, _ := types.ListValueFrom(context.Background(), types.StringType, names)
	m.RoleNames = listVal
}
