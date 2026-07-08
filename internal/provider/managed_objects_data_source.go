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

var _ datasource.DataSource = &managedObjectsDataSource{}

type managedObjectsDataSource struct {
	client *client.Client
}

func NewManagedObjectsDataSource() datasource.DataSource {
	return &managedObjectsDataSource{}
}

type managedObjectsDataModel struct {
	Type           types.String `tfsdk:"type"`
	FragmentType   types.String `tfsdk:"fragment_type"`
	Query          types.String `tfsdk:"query"`
	Text           types.String `tfsdk:"text"`
	Owner          types.String `tfsdk:"owner"`
	ManagedObjects types.List   `tfsdk:"managed_objects"`
}

var managedObjectAttrTypes = map[string]attr.Type{
	"id":              types.StringType,
	"name":            types.StringType,
	"type":            types.StringType,
	"owner":           types.StringType,
	"self":            types.StringType,
	"creation_time":   types.StringType,
	"last_updated":    types.StringType,
	"is_device":       types.BoolType,
	"is_device_group": types.BoolType,
}

func (d *managedObjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_objects"
}

func (d *managedObjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of managed objects (devices, assets, groups, etc.) from the Cumulocity inventory, " +
			"optionally filtered by type, fragment, query, text prefix, or owner. All pages are followed automatically. " +
			"Corresponds to GET /inventory/managedObjects.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by managed object type, e.g. `c8y_MQTTDevice`.",
			},
			"fragment_type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the presence of a specific fragment, e.g. `c8y_IsDevice`.",
			},
			"query": schema.StringAttribute{
				Optional:    true,
				Description: "Advanced inventory query string (Cumulocity query language).",
			},
			"text": schema.StringAttribute{
				Optional:    true,
				Description: "Full-text search — returns objects whose name contains this string.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by owner username.",
			},
			"managed_objects": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching managed objects.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "Managed object ID."},
						"name":            schema.StringAttribute{Computed: true, Description: "Name of the managed object."},
						"type":            schema.StringAttribute{Computed: true, Description: "Type of the managed object."},
						"owner":           schema.StringAttribute{Computed: true, Description: "Owner username."},
						"self":            schema.StringAttribute{Computed: true, Description: "Self-link URL."},
						"creation_time":   schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"last_updated":    schema.StringAttribute{Computed: true, Description: "ISO 8601 last update timestamp."},
						"is_device":       schema.BoolAttribute{Computed: true, Description: "Whether the object carries `c8y_IsDevice`."},
						"is_device_group": schema.BoolAttribute{Computed: true, Description: "Whether the object carries `c8y_IsDeviceGroup`."},
					},
				},
			},
		},
	}
}

func (d *managedObjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *managedObjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config managedObjectsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects, err := d.client.ListManagedObjects(ctx,
		config.Type.ValueString(),
		config.FragmentType.ValueString(),
		config.Query.ValueString(),
		config.Text.ValueString(),
		config.Owner.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing managed objects", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(objects))
	for _, mo := range objects {
		obj, diags := types.ObjectValue(managedObjectAttrTypes, map[string]attr.Value{
			"id":              types.StringValue(mo.ID),
			"name":            types.StringValue(mo.Name),
			"type":            types.StringValue(mo.Type),
			"owner":           types.StringValue(mo.Owner),
			"self":            types.StringValue(mo.Self),
			"creation_time":   types.StringValue(mo.CreationTime),
			"last_updated":    types.StringValue(mo.LastUpdated),
			"is_device":       types.BoolValue(mo.C8yIsDevice != nil),
			"is_device_group": types.BoolValue(mo.C8yIsDeviceGroup != nil),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: managedObjectAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ManagedObjects = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
