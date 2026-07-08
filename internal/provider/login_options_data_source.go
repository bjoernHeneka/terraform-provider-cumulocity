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

var _ datasource.DataSource = &loginOptionsDataSource{}

type loginOptionsDataSource struct {
	client *client.Client
}

func NewLoginOptionsDataSource() datasource.DataSource {
	return &loginOptionsDataSource{}
}

type loginOptionsModel struct {
	Options types.List `tfsdk:"options"`
}

var loginOptionAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"type":                   types.StringType,
	"provider_name":          types.StringType,
	"grant_type":             types.StringType,
	"user_management_source": types.StringType,
	"visible_on_login_page":  types.BoolType,
	"template":               types.StringType,
	"button_name":            types.StringType,
	"issuer":                 types.StringType,
	"client_id":              types.StringType,
	"audience":               types.StringType,
	"redirect_to_platform":   types.StringType,
	"use_id_token":           types.BoolType,
	"self":                   types.StringType,
	"config_json":            types.StringType,
}

func (d *loginOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login_options"
}

func (d *loginOptionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all login options configured on the tenant. " +
			"Corresponds to GET /tenant/loginOptions.",
		Attributes: map[string]schema.Attribute{
			"options": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of all login options on the tenant.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The login option ID.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The authentication configuration type (e.g. OAUTH2_INTERNAL, BASIC).",
						},
						"provider_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the authentication provider.",
						},
						"grant_type": schema.StringAttribute{
							Computed:    true,
							Description: "The OAuth2 grant type (e.g. PASSWORD, AUTHORIZATION_CODE).",
						},
						"user_management_source": schema.StringAttribute{
							Computed:    true,
							Description: "The source of user management (e.g. INTERNAL).",
						},
						"visible_on_login_page": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this login option is shown on the login page.",
						},
						"template": schema.StringAttribute{
							Computed:    true,
							Description: "The configuration template, e.g. CUSTOM (OAuth2 options).",
						},
						"button_name": schema.StringAttribute{
							Computed:    true,
							Description: "The label of the login button shown on the login page.",
						},
						"issuer": schema.StringAttribute{
							Computed:    true,
							Description: "The OAuth2/OIDC token issuer URL.",
						},
						"client_id": schema.StringAttribute{
							Computed:    true,
							Description: "The OAuth2 client ID.",
						},
						"audience": schema.StringAttribute{
							Computed:    true,
							Description: "The OAuth2 token audience.",
						},
						"redirect_to_platform": schema.StringAttribute{
							Computed:    true,
							Description: "The platform redirect URL used in the OAuth2 flow.",
						},
						"use_id_token": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the ID token is used instead of the access token.",
						},
						"self": schema.StringAttribute{
							Computed:    true,
							Description: "The self-link URL of the login option.",
						},
						"config_json": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
							Description: "The complete raw JSON payload as returned by the API, including " +
								"all type-specific nested fields. Parse with jsondecode(). Marked " +
								"sensitive as request templates may contain secrets.",
						},
					},
				},
			},
		},
	}
}

func (d *loginOptionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *loginOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loginOptionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts, err := d.client.ListLoginOptions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing login options", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(opts))
	for _, opt := range opts {
		obj, diags := types.ObjectValue(loginOptionAttrTypes, map[string]attr.Value{
			"id":                     types.StringValue(opt.ID),
			"type":                   types.StringValue(opt.Type),
			"provider_name":          types.StringValue(opt.ProviderName),
			"grant_type":             types.StringValue(opt.GrantType),
			"user_management_source": types.StringValue(opt.UserManagementSource),
			"visible_on_login_page":  types.BoolValue(opt.VisibleOnLoginPage),
			"template":               types.StringValue(opt.Template),
			"button_name":            types.StringValue(opt.ButtonName),
			"issuer":                 types.StringValue(opt.Issuer),
			"client_id":              types.StringValue(opt.ClientID),
			"audience":               types.StringValue(opt.Audience),
			"redirect_to_platform":   types.StringValue(opt.RedirectToPlatform),
			"use_id_token":           types.BoolValue(opt.UseIDToken),
			"self":                   types.StringValue(opt.Self),
			"config_json":            types.StringValue(string(opt.Raw)),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: loginOptionAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Options = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
