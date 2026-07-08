package provider

import (
	"context"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &loginOptionDataSource{}

type loginOptionDataSource struct {
	client *client.Client
}

func NewLoginOptionDataSource() datasource.DataSource {
	return &loginOptionDataSource{}
}

type loginOptionDataSourceModel struct {
	// Input
	TypeOrID types.String `tfsdk:"type_or_id"`
	// Computed
	ID                   types.String `tfsdk:"id"`
	Type                 types.String `tfsdk:"type"`
	ProviderName         types.String `tfsdk:"provider_name"`
	GrantType            types.String `tfsdk:"grant_type"`
	UserManagementSource types.String `tfsdk:"user_management_source"`
	VisibleOnLoginPage   types.Bool   `tfsdk:"visible_on_login_page"`
	Template             types.String `tfsdk:"template"`
	ButtonName           types.String `tfsdk:"button_name"`
	Issuer               types.String `tfsdk:"issuer"`
	ClientID             types.String `tfsdk:"client_id"`
	Audience             types.String `tfsdk:"audience"`
	RedirectToPlatform   types.String `tfsdk:"redirect_to_platform"`
	UseIDToken           types.Bool   `tfsdk:"use_id_token"`
	Self                 types.String `tfsdk:"self"`
	ConfigJSON           types.String `tfsdk:"config_json"`
}

func (d *loginOptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login_option"
}

func (d *loginOptionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single Cumulocity login option by its type or ID. " +
			"Corresponds to GET /tenant/loginOptions/{typeOrId}.",
		Attributes: map[string]schema.Attribute{
			"type_or_id": schema.StringAttribute{
				Required:    true,
				Description: "The login option type (e.g. OAUTH2_INTERNAL) or its ID to look up.",
			},
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
				Description: "The complete raw JSON payload as returned by the API, including all " +
					"type-specific nested fields (tokenRequest, authorizationRequest, onNewUser, " +
					"signatureVerificationConfig, etc.). Parse with jsondecode(). Marked sensitive " +
					"as request templates may contain secrets.",
			},
		},
	}
}

func (d *loginOptionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *loginOptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config loginOptionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt, err := d.client.GetLoginOption(ctx, config.TypeOrID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading login option", err.Error())
		return
	}

	config.ID = types.StringValue(opt.ID)
	config.Type = types.StringValue(opt.Type)
	config.ProviderName = types.StringValue(opt.ProviderName)
	config.GrantType = types.StringValue(opt.GrantType)
	config.UserManagementSource = types.StringValue(opt.UserManagementSource)
	config.VisibleOnLoginPage = types.BoolValue(opt.VisibleOnLoginPage)
	config.Template = types.StringValue(opt.Template)
	config.ButtonName = types.StringValue(opt.ButtonName)
	config.Issuer = types.StringValue(opt.Issuer)
	config.ClientID = types.StringValue(opt.ClientID)
	config.Audience = types.StringValue(opt.Audience)
	config.RedirectToPlatform = types.StringValue(opt.RedirectToPlatform)
	config.UseIDToken = types.BoolValue(opt.UseIDToken)
	config.Self = types.StringValue(opt.Self)
	config.ConfigJSON = types.StringValue(string(opt.Raw))

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
