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

var _ datasource.DataSource = &roleDataSource{}

type roleDataSource struct {
	client *client.Client
}

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSourceModel struct {
	// Input
	Name types.String `tfsdk:"name"`
	// Computed
	ID   types.String `tfsdk:"id"`
	Self types.String `tfsdk:"self"`
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single Cumulocity global role by name. " +
			"Useful for referencing roles in cumulocity_user_role_assignment resources. " +
			"Corresponds to GET /user/roles/{name}.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The role name, e.g. \"ROLE_ALARM_ADMIN\".",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The role identifier (equals name in Cumulocity).",
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "The self-link URL of the role.",
			},
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config roleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(ctx, config.Name.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Role not found",
			fmt.Sprintf("No role with name %q exists in Cumulocity.", config.Name.ValueString()),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	config.ID = types.StringValue(role.ID)
	config.Self = types.StringValue(role.Self)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
