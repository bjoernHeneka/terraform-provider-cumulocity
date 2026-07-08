package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ datasource.DataSource = &rolesDataSource{}

type rolesDataSource struct {
	client *client.Client
}

func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

// rolesDataSourceModel is the full state for the cumulocity_roles data source.
type rolesDataSourceModel struct {
	// Optional filter
	NameFilter types.String `tfsdk:"name_filter"`
	// Computed: list of matched roles
	Roles []roleItemModel `tfsdk:"roles"`
}

// roleItemModel is one entry in the roles list.
type roleItemModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Self types.String `tfsdk:"self"`
}

// roleItemAttrTypes maps the tfsdk attribute names to their framework types.
// Used when building types.List values.
var roleItemAttrTypes = map[string]attr.Type{
	"id":   types.StringType,
	"name": types.StringType,
	"self": types.StringType,
}

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all available global roles in Cumulocity. " +
			"Follows pagination automatically to return the complete list. " +
			"Corresponds to GET /user/roles.",
		Attributes: map[string]schema.Attribute{
			"name_filter": schema.StringAttribute{
				Optional:    true,
				Description: "When set, only roles whose name contains this string (case-insensitive) are returned.",
			},
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of matching roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The role identifier (equals name in Cumulocity).",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The role name, e.g. \"ROLE_ALARM_ADMIN\".",
						},
						"self": schema.StringAttribute{
							Computed:    true,
							Description: "The self-link URL of the role.",
						},
					},
				},
			},
		},
	}
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config rolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roles, err := d.client.ListRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing roles", err.Error())
		return
	}

	filter := config.NameFilter.ValueString()

	var items []roleItemModel
	for _, r := range roles {
		if filter != "" && !containsFold(r.Name, filter) {
			continue
		}
		items = append(items, roleItemModel{
			ID:   types.StringValue(r.ID),
			Name: types.StringValue(r.Name),
			Self: types.StringValue(r.Self),
		})
	}

	if items == nil {
		items = []roleItemModel{} // never return null — always an empty list
	}

	config.Roles = items
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

// containsFold reports whether s contains substr, case-insensitively.
func containsFold(s, substr string) bool {
	return len(s) >= len(substr) &&
		foldContains([]rune(s), []rune(substr))
}

func foldContains(s, sub []rune) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		match := true
		for j, r := range sub {
			if foldRune(s[i+j]) != foldRune(r) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func foldRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}
