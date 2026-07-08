package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ datasource.DataSource = &inventoryRoleDataSource{}

type inventoryRoleDataSource struct {
	client *client.Client
}

func NewInventoryRoleDataSource() datasource.DataSource {
	return &inventoryRoleDataSource{}
}

type inventoryRoleDataSourceModel struct {
	// Exactly one of name or id must be set as input.
	Name types.String `tfsdk:"name"`
	ID   types.Int64  `tfsdk:"id"`

	// Computed
	Description types.String               `tfsdk:"description"`
	Self        types.String               `tfsdk:"self"`
	Permissions []inventoryPermissionModel `tfsdk:"permissions"`
}

type inventoryPermissionModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Permission types.String `tfsdk:"permission"`
	Scope      types.String `tfsdk:"scope"`
	Type       types.String `tfsdk:"type"`
}

func (d *inventoryRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory_role"
}

func (d *inventoryRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	permissionAttrs := map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:    true,
			Description: "Unique identifier of the permission entry.",
		},
		"permission": schema.StringAttribute{
			Computed:    true,
			Description: "Permission level: ADMIN, READ, or *.",
		},
		"scope": schema.StringAttribute{
			Computed:    true,
			Description: "Scope: ALARM, AUDIT, EVENT, MANAGED_OBJECT, MEASUREMENT, OPERATION, or *.",
		},
		"type": schema.StringAttribute{
			Computed:    true,
			Description: "Fragment type filter, e.g. \"c8y_Restart\". Empty string means all types.",
		},
	}

	resp.Schema = schema.Schema{
		Description: "Looks up a single Cumulocity inventory role by name or numeric ID. " +
			"Provide exactly one of name or id. " +
			"Corresponds to GET /user/inventoryroles and GET /user/inventoryroles/{id}.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exact name of the inventory role, e.g. \"Operations: Restart Device\". Provide name or id.",
			},
			"id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Numeric ID of the inventory role. Provide name or id.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the inventory role.",
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the inventory role.",
			},
			"permissions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of permissions defined for this inventory role.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: permissionAttrs,
				},
			},
		},
	}
}

func (d *inventoryRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *inventoryRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config inventoryRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""
	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()

	if !hasName && !hasID {
		resp.Diagnostics.AddError(
			"Missing lookup key",
			"Provide either name or id to look up an inventory role.",
		)
		return
	}

	var role *client.InventoryRole
	var err error

	if hasID {
		role, err = d.client.GetInventoryRole(ctx, config.ID.ValueInt64())
	} else {
		role, err = d.client.GetInventoryRoleByName(ctx, config.Name.ValueString())
	}

	if errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Inventory role not found",
			err.Error(),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading inventory role", err.Error())
		return
	}

	config.ID = types.Int64Value(role.ID)
	config.Name = types.StringValue(role.Name)
	config.Description = types.StringValue(role.Description)
	config.Self = types.StringValue(role.Self)
	config.Permissions = permissionsToModel(role.Permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func permissionsToModel(perms []client.InventoryRolePermission) []inventoryPermissionModel {
	result := make([]inventoryPermissionModel, len(perms))
	for i, p := range perms {
		result[i] = inventoryPermissionModel{
			ID:         types.Int64Value(p.ID),
			Permission: types.StringValue(p.Permission),
			Scope:      types.StringValue(p.Scope),
			Type:       types.StringValue(p.Type),
		}
	}
	return result
}
