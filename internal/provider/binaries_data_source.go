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

var _ datasource.DataSource = &binariesDataSource{}

type binariesDataSource struct {
	client *client.Client
}

func NewBinariesDataSource() datasource.DataSource {
	return &binariesDataSource{}
}

type binariesDataModel struct {
	Owner    types.String `tfsdk:"owner"`
	Type     types.String `tfsdk:"type"`
	Binaries types.List   `tfsdk:"binaries"`
}

var binaryItemAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"type":         types.StringType,
	"content_type": types.StringType,
	"length":       types.Int64Type,
	"owner":        types.StringType,
	"self":         types.StringType,
	"last_updated": types.StringType,
}

func (d *binariesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_binaries"
}

func (d *binariesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves metadata for files stored in the Cumulocity inventory binary store, " +
			"optionally filtered by owner or type. All pages are followed automatically. " +
			"Corresponds to GET /inventory/binaries.",
		Attributes: map[string]schema.Attribute{
			"owner": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the username of the binary owner.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the managed object type of the binary.",
			},
			"binaries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching binary managed objects.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "Binary managed object ID."},
						"name":         schema.StringAttribute{Computed: true, Description: "File name."},
						"type":         schema.StringAttribute{Computed: true, Description: "Managed object type."},
						"content_type": schema.StringAttribute{Computed: true, Description: "MIME content type of the stored file."},
						"length":       schema.Int64Attribute{Computed: true, Description: "File size in bytes."},
						"owner":        schema.StringAttribute{Computed: true, Description: "Username of the owner."},
						"self":         schema.StringAttribute{Computed: true, Description: "Self-link URL."},
						"last_updated": schema.StringAttribute{Computed: true, Description: "ISO 8601 timestamp of the last update."},
					},
				},
			},
		},
	}
}

func (d *binariesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *binariesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config binariesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	binaries, err := d.client.ListBinaries(ctx,
		config.Owner.ValueString(),
		config.Type.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing binaries", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(binaries))
	for _, b := range binaries {
		obj, diags := types.ObjectValue(binaryItemAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(b.ID),
			"name":         types.StringValue(b.Name),
			"type":         types.StringValue(b.Type),
			"content_type": types.StringValue(b.ContentType),
			"length":       types.Int64Value(b.Length),
			"owner":        types.StringValue(b.Owner),
			"self":         types.StringValue(b.Self),
			"last_updated": types.StringValue(b.LastUpdated),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: binaryItemAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Binaries = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
