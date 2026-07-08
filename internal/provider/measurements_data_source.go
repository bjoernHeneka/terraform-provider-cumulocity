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

var _ datasource.DataSource = &measurementsDataSource{}

type measurementsDataSource struct {
	client *client.Client
}

func NewMeasurementsDataSource() datasource.DataSource {
	return &measurementsDataSource{}
}

type measurementsDataModel struct {
	SourceID            types.String `tfsdk:"source_id"`
	Type                types.String `tfsdk:"type"`
	DateFrom            types.String `tfsdk:"date_from"`
	DateTo              types.String `tfsdk:"date_to"`
	ValueFragmentType   types.String `tfsdk:"value_fragment_type"`
	ValueFragmentSeries types.String `tfsdk:"value_fragment_series"`
	Measurements        types.List   `tfsdk:"measurements"`
}

var measurementAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"source_id":     types.StringType,
	"type":          types.StringType,
	"time":          types.StringType,
	"creation_time": types.StringType,
	"self":          types.StringType,
}

func (d *measurementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_measurements"
}

func (d *measurementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of measurements from the Cumulocity measurement API, " +
			"optionally filtered by source device, type, date range, or fragment series. " +
			"All pages are followed automatically. " +
			"Corresponds to GET /measurement/measurements.",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the source managed object (device) ID.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by measurement type, e.g. `c8y_TemperatureMeasurement`.",
			},
			"date_from": schema.StringAttribute{
				Optional:    true,
				Description: "Start of date range (ISO 8601).",
			},
			"date_to": schema.StringAttribute{
				Optional:    true,
				Description: "End of date range (ISO 8601).",
			},
			"value_fragment_type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the fragment type that contains the measurement value, e.g. `c8y_Steam`.",
			},
			"value_fragment_series": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the series name within the fragment, e.g. `Temperature`.",
			},
			"measurements": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching measurements.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, Description: "Measurement ID."},
						"source_id":     schema.StringAttribute{Computed: true, Description: "Source managed object ID."},
						"type":          schema.StringAttribute{Computed: true, Description: "Measurement type."},
						"time":          schema.StringAttribute{Computed: true, Description: "ISO 8601 time the measurement was taken."},
						"creation_time": schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"self":          schema.StringAttribute{Computed: true, Description: "Self-link URL."},
					},
				},
			},
		},
	}
}

func (d *measurementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *measurementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config measurementsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	measurements, err := d.client.ListMeasurements(ctx,
		config.SourceID.ValueString(),
		config.Type.ValueString(),
		config.DateFrom.ValueString(),
		config.DateTo.ValueString(),
		config.ValueFragmentType.ValueString(),
		config.ValueFragmentSeries.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing measurements", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(measurements))
	for _, m := range measurements {
		obj, diags := types.ObjectValue(measurementAttrTypes, map[string]attr.Value{
			"id":            types.StringValue(m.ID),
			"source_id":     types.StringValue(m.Source.ID),
			"type":          types.StringValue(m.Type),
			"time":          types.StringValue(m.Time),
			"creation_time": types.StringValue(m.CreationTime),
			"self":          types.StringValue(m.Self),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: measurementAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Measurements = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
