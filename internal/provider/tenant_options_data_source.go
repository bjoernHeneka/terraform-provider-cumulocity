package provider

import (
	"context"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &tenantOptionsDataSource{}

type tenantOptionsDataSource struct {
	client *client.Client
}

func NewTenantOptionsDataSource() datasource.DataSource {
	return &tenantOptionsDataSource{}
}

type tenantOptionsModel struct {
	Category types.String `tfsdk:"category"`
	Options  types.List   `tfsdk:"options"`
}

var tenantOptionAttrTypes = map[string]attr.Type{
	"category": types.StringType,
	"key":      types.StringType,
	"value":    types.StringType,
	"self":     types.StringType,
}

func (d *tenantOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_options"
}

func (d *tenantOptionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves all tenant options, optionally filtered by category. " +
			"Corresponds to GET /tenant/options (paginated).",
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Optional:    true,
				Description: "When set, only options belonging to this category are returned.",
			},
			"options": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of tenant options matching the filter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"category": schema.StringAttribute{
							Computed:    true,
							Description: "Category of the option.",
						},
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "Key of the option.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "Value of the option. Marked sensitive as it may contain credentials.",
						},
						"self": schema.StringAttribute{
							Computed:    true,
							Description: "Self-link URL of the option.",
						},
					},
				},
			},
		},
	}
}

func (d *tenantOptionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tenantOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tenantOptionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	category := config.Category.ValueString()
	opts, err := d.client.ListTenantOptions(ctx, category)
	if err != nil {
		resp.Diagnostics.AddError("Error listing tenant options", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(opts))
	for _, opt := range opts {
		obj, diags := types.ObjectValue(tenantOptionAttrTypes, map[string]attr.Value{
			"category": types.StringValue(opt.Category),
			"key":      types.StringValue(opt.Key),
			"value":    types.StringValue(opt.Value),
			"self":     types.StringValue(opt.Self),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: tenantOptionAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Options = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
