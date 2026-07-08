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

var _ datasource.DataSource = &alarmsDataSource{}

type alarmsDataSource struct {
	client *client.Client
}

func NewAlarmsDataSource() datasource.DataSource {
	return &alarmsDataSource{}
}

type alarmsDataModel struct {
	SourceID types.String `tfsdk:"source_id"`
	Status   types.String `tfsdk:"status"`
	Severity types.String `tfsdk:"severity"`
	Type     types.String `tfsdk:"type"`
	Alarms   types.List   `tfsdk:"alarms"`
}

var alarmAttrTypes = map[string]attr.Type{
	"id":                    types.StringType,
	"source_id":             types.StringType,
	"type":                  types.StringType,
	"text":                  types.StringType,
	"severity":              types.StringType,
	"status":                types.StringType,
	"time":                  types.StringType,
	"occurrence_count":      types.Int64Type,
	"creation_time":         types.StringType,
	"last_updated":          types.StringType,
	"first_occurrence_time": types.StringType,
}

func (d *alarmsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alarms"
}

func (d *alarmsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of alarms, optionally filtered by source device, status, severity, and/or type. " +
			"All pages are followed automatically. " +
			"Corresponds to GET /alarm/alarms.",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter alarms by the managed object (device) ID.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by alarm status: ACTIVE, ACKNOWLEDGED, or CLEARED.",
			},
			"severity": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by alarm severity: CRITICAL, MAJOR, MINOR, or WARNING.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by alarm type, e.g. `c8y_UnavailabilityAlarm`.",
			},
			"alarms": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching alarms.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true, Description: "Alarm ID."},
						"source_id":             schema.StringAttribute{Computed: true, Description: "Source managed object ID."},
						"type":                  schema.StringAttribute{Computed: true, Description: "Alarm type."},
						"text":                  schema.StringAttribute{Computed: true, Description: "Alarm description."},
						"severity":              schema.StringAttribute{Computed: true, Description: "Severity: CRITICAL, MAJOR, MINOR, or WARNING."},
						"status":                schema.StringAttribute{Computed: true, Description: "Status: ACTIVE, ACKNOWLEDGED, or CLEARED."},
						"time":                  schema.StringAttribute{Computed: true, Description: "ISO 8601 time the alarm occurred."},
						"occurrence_count":      schema.Int64Attribute{Computed: true, Description: "Number of times this alarm was triggered."},
						"creation_time":         schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"last_updated":          schema.StringAttribute{Computed: true, Description: "ISO 8601 last update timestamp."},
						"first_occurrence_time": schema.StringAttribute{Computed: true, Description: "ISO 8601 first occurrence timestamp (set when count > 1)."},
					},
				},
			},
		},
	}
}

func (d *alarmsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alarmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alarmsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alarms, err := d.client.ListAlarms(ctx,
		config.SourceID.ValueString(),
		config.Status.ValueString(),
		config.Severity.ValueString(),
		config.Type.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing alarms", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(alarms))
	for _, a := range alarms {
		obj, diags := types.ObjectValue(alarmAttrTypes, map[string]attr.Value{
			"id":                    types.StringValue(a.ID),
			"source_id":             types.StringValue(a.Source.ID),
			"type":                  types.StringValue(a.Type),
			"text":                  types.StringValue(a.Text),
			"severity":              types.StringValue(a.Severity),
			"status":                types.StringValue(a.Status),
			"time":                  types.StringValue(a.Time),
			"occurrence_count":      types.Int64Value(a.Count),
			"creation_time":         types.StringValue(a.CreationTime),
			"last_updated":          types.StringValue(a.LastUpdated),
			"first_occurrence_time": types.StringValue(a.FirstOccurrenceTime),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: alarmAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Alarms = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
