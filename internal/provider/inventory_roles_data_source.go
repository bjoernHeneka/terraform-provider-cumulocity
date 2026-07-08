package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ datasource.DataSource = &inventoryRolesDataSource{}

type inventoryRolesDataSource struct {
	client *client.Client
}

func NewInventoryRolesDataSource() datasource.DataSource {
	return &inventoryRolesDataSource{}
}

type inventoryRolesDataSourceModel struct {
	NameFilter types.String             `tfsdk:"name_filter"`
	Roles      []inventoryRoleItemModel `tfsdk:"roles"`
}

type inventoryRoleItemModel struct {
	ID          types.Int64                `tfsdk:"id"`
	Name        types.String               `tfsdk:"name"`
	Description types.String               `tfsdk:"description"`
	Self        types.String               `tfsdk:"self"`
	Permissions []inventoryPermissionModel `tfsdk:"permissions"`
}

func (d *inventoryRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_inventory_roles"
}

func (d *inventoryRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
		Description: "Lists all Cumulocity inventory roles with their permissions. " +
			"Follows pagination automatically. " +
			"Corresponds to GET /user/inventoryroles.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Optional:    true,
				Description: "When set, only roles whose name contains this string (case-insensitive) are returned.",
			},
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of matching inventory roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "Numeric ID of the inventory role.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the inventory role.",
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
							Description: "Permissions defined for this inventory role.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: permissionAttrs,
							},
						},
					},
				},
			},
		},
	}
}

func (d *inventoryRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *inventoryRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config inventoryRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := d.client.ListInventoryRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing inventory roles", err.Error())
		return
	}

	filter := config.NameFilter.ValueString()

	var items []inventoryRoleItemModel
	for _, r := range roles {
		if filter != "" && !containsFold(r.Name, filter) {
			continue
		}
		items = append(items, inventoryRoleItemModel{
			ID:          types.Int64Value(r.ID),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
			Self:        types.StringValue(r.Self),
			Permissions: permissionsToModel(r.Permissions),
		})
	}

	if items == nil {
		items = []inventoryRoleItemModel{}
	}

	config.Roles = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
