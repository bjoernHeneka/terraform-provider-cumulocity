package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ datasource.DataSource = &applicationDataSource{}

type applicationDataSource struct {
	client *client.Client
}

func NewApplicationDataSource() datasource.DataSource {
	return &applicationDataSource{}
}

type applicationDataSourceModel struct {
	// Input
	Name types.String `tfsdk:"name"`
	// Computed
	ID              types.String `tfsdk:"id"`
	Key             types.String `tfsdk:"key"`
	Type            types.String `tfsdk:"type"`
	ContextPath     types.String `tfsdk:"context_path"`
	Availability    types.String `tfsdk:"availability"`
	Description     types.String `tfsdk:"description"`
	ActiveVersionID types.String `tfsdk:"active_version_id"`
	OwnerTenantID   types.String `tfsdk:"owner_tenant_id"`
	Self            types.String `tfsdk:"self"`
}

func (d *applicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *applicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single Cumulocity application by name. " +
			"Useful for referencing applications in other resources like cumulocity_tenant_application_subscription. " +
			"Corresponds to GET /application/applicationsByName/{name}.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The application name to look up.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The application ID.",
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Description: "The application key.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Application type: HOSTED, EXTERNAL, or MICROSERVICE.",
			},
			"context_path": schema.StringAttribute{
				Computed:    true,
				Description: "The application context path (for HOSTED applications).",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "Application availability: MARKET or PRIVATE.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Application description.",
			},
			"active_version_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the active binary version (set after upload).",
			},
			"owner_tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the tenant that owns this application.",
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "The self-link URL of the application.",
			},
		},
	}
}

func (d *applicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *applicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config applicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apps, err := d.client.GetApplicationsByName(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	if len(apps) == 0 {
		resp.Diagnostics.AddError(
			"Application not found",
			fmt.Sprintf("No application with name %q exists in Cumulocity.", config.Name.ValueString()),
		)
		return
	}

	if len(apps) > 1 {
		resp.Diagnostics.AddError(
			"Multiple applications found",
			fmt.Sprintf("Found %d applications with name %q. Application names should be unique for data source lookups.", len(apps), config.Name.ValueString()),
		)
		return
	}

	app := apps[0]
	config.ID = types.StringValue(app.ID)
	config.Key = types.StringValue(app.Key)
	config.Type = types.StringValue(app.Type)
	config.ContextPath = types.StringValue(app.ContextPath)
	config.Availability = types.StringValue(app.Availability)
	config.Description = types.StringValue(app.Description)
	config.ActiveVersionID = types.StringValue(app.ActiveVersionID)
	config.Self = types.StringValue(app.Self)

	if app.Owner != nil {
		config.OwnerTenantID = types.StringValue(app.Owner.Tenant.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
