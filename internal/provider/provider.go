package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Compile-time interface check.
var _ provider.Provider = &cumulocityProvider{}

type cumulocityProvider struct {
	version string
}

// New returns a provider factory function used by main.go.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cumulocityProvider{version: version}
	}
}

// cumulocityProviderModel holds the provider block configuration.
type cumulocityProviderModel struct {
	TenantDomain types.String `tfsdk:"tenant_domain"`
	TenantID     types.String `tfsdk:"tenant_id"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
}

func (p *cumulocityProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cumulocity"
	resp.Version = p.version
}

func (p *cumulocityProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing Cumulocity IoT platform resources via the Cumulocity REST API.",
		Attributes: map[string]schema.Attribute{
			"tenant_domain": schema.StringAttribute{
				Optional:    true,
				Description: "The Cumulocity tenant domain (e.g. \"mytenant.cumulocity.com\"). Can also be set via the CUMULOCITY_TENANT_DOMAIN environment variable.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Cumulocity tenant ID (short ID, e.g. \"t0071234\"). Used for Basic auth credential construction: base64(<tenantID>/<username>:<password>). Can also be set via the CUMULOCITY_TENANT_ID environment variable.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for Cumulocity Basic auth. Can also be set via the CUMULOCITY_USERNAME environment variable.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for Cumulocity Basic auth. Can also be set via the CUMULOCITY_PASSWORD environment variable.",
			},
		},
	}
}

func (p *cumulocityProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config cumulocityProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unknown values cannot be resolved at configuration time (they depend on
	// another resource's output). Fail early with a clear pointer to env vars.
	if config.TenantDomain.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_domain"),
			"Unknown tenant_domain",
			"tenant_domain cannot be an unknown value at provider configuration time. "+
				"Set it to a known value or use the CUMULOCITY_TENANT_DOMAIN environment variable.",
		)
	}
	if config.TenantID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Unknown tenant_id",
			"tenant_id cannot be an unknown value at provider configuration time. "+
				"Set it to a known value or use the CUMULOCITY_TENANT_ID environment variable.",
		)
	}
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown username",
			"username cannot be an unknown value at provider configuration time. "+
				"Set it to a known value or use the CUMULOCITY_USERNAME environment variable.",
		)
	}
	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown password",
			"password cannot be an unknown value at provider configuration time. "+
				"Set it to a known value or use the CUMULOCITY_PASSWORD environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Environment variables as fallback; explicit config block values take precedence.
	tenantDomain := os.Getenv("CUMULOCITY_TENANT_DOMAIN")
	if !config.TenantDomain.IsNull() && !config.TenantDomain.IsUnknown() {
		tenantDomain = config.TenantDomain.ValueString()
	}

	tenantID := os.Getenv("CUMULOCITY_TENANT_ID")
	if !config.TenantID.IsNull() && !config.TenantID.IsUnknown() {
		tenantID = config.TenantID.ValueString()
	}

	username := os.Getenv("CUMULOCITY_USERNAME")
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}

	password := os.Getenv("CUMULOCITY_PASSWORD")
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}

	if tenantDomain == "" {
		resp.Diagnostics.AddError(
			"Missing tenant_domain",
			"Set tenant_domain in the provider block or the CUMULOCITY_TENANT_DOMAIN environment variable.",
		)
	}
	if username == "" {
		resp.Diagnostics.AddError(
			"Missing username",
			"Set username in the provider block or the CUMULOCITY_USERNAME environment variable.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddError(
			"Missing password",
			"Set password in the provider block or the CUMULOCITY_PASSWORD environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.NewClient(tenantDomain, tenantID, username, password)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cumulocity API client", err.Error())
		return
	}
	c.SetUserAgent(fmt.Sprintf("terraform-provider-cumulocity/%s", p.version))

	// Validate credentials up front so misconfiguration fails during configure
	// rather than on the first resource operation. The error is intentionally
	// credential-free.
	if _, err := c.GetCurrentUser(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to authenticate with Cumulocity",
			"Verify tenant_domain, tenant_id, username and password (or the corresponding "+
				"CUMULOCITY_* environment variables) and that the tenant is reachable.",
		)
		return
	}

	// Store the client so resources and data sources can retrieve it.
	resp.ResourceData = c
	resp.DataSourceData = c
}

// Resources returns the list of implemented resource types.
// Add new resources here as they are implemented.
func (p *cumulocityProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewUserRoleAssignmentResource,
		NewUserInventoryRoleAssignmentResource,
		NewManagedObjectResource,
		NewUserGroupResource,
		NewUserGroupMembershipResource,
		NewTenantOptionResource,
		NewApplicationResource,
		NewApplicationBinaryResource,
		NewDeviceOperationResource,
		NewNewDeviceRequestResource,
		NewDeviceCredentialsResource,
		NewBulkOperationResource,
		NewAlarmResource,
		NewAuditRecordResource,
		NewEventResource,
		NewTenantResource,
		NewTenantApplicationSubscriptionResource,
		NewTrustedCertificateResource,
		NewLoginOptionResource,
		NewLoginOptionRawResource,
		NewExternalIDResource,
		NewBinaryResource,
		NewMeasurementResource,
		NewRetentionRuleResource,
		NewNotificationSubscriptionResource,
	}
}

// DataSources returns the list of implemented data source types.
func (p *cumulocityProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRoleDataSource,
		NewRolesDataSource,
		NewInventoryRoleDataSource,
		NewInventoryRolesDataSource,
		NewTenantOptionsDataSource,
		NewOperationsDataSource,
		NewAlarmsDataSource,
		NewAuditRecordsDataSource,
		NewEventsDataSource,
		NewManagedObjectsDataSource,
		NewMeasurementsDataSource,
		NewBinariesDataSource,
		NewApplicationDataSource,
		NewLoginOptionDataSource,
		NewLoginOptionsDataSource,
	}
}
