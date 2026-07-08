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

var _ datasource.DataSource = &eventsDataSource{}

type eventsDataSource struct {
	client *client.Client
}

func NewEventsDataSource() datasource.DataSource {
	return &eventsDataSource{}
}

type eventsDataModel struct {
	SourceID types.String `tfsdk:"source_id"`
	Type     types.String `tfsdk:"type"`
	DateFrom types.String `tfsdk:"date_from"`
	DateTo   types.String `tfsdk:"date_to"`
	Events   types.List   `tfsdk:"events"`
}

var eventAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"source_id":     types.StringType,
	"type":          types.StringType,
	"text":          types.StringType,
	"time":          types.StringType,
	"creation_time": types.StringType,
	"last_updated":  types.StringType,
}

func (d *eventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_events"
}

func (d *eventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of events, optionally filtered by source device, type, and date range. " +
			"All pages are followed automatically. " +
			"Corresponds to GET /event/events.",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter events by the managed object (device/asset) ID.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by event type, e.g. `c8y_LocationUpdate`.",
			},
			"date_from": schema.StringAttribute{
				Optional:    true,
				Description: "Start of date range (ISO 8601). Filters by device timestamp.",
			},
			"date_to": schema.StringAttribute{
				Optional:    true,
				Description: "End of date range (ISO 8601). Filters by device timestamp.",
			},
			"events": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching events.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, Description: "Event ID."},
						"source_id":     schema.StringAttribute{Computed: true, Description: "Source managed object ID."},
						"type":          schema.StringAttribute{Computed: true, Description: "Event type."},
						"text":          schema.StringAttribute{Computed: true, Description: "Event description."},
						"time":          schema.StringAttribute{Computed: true, Description: "ISO 8601 time the event occurred."},
						"creation_time": schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"last_updated":  schema.StringAttribute{Computed: true, Description: "ISO 8601 last update timestamp."},
					},
				},
			},
		},
	}
}

func (d *eventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *eventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config eventsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	events, err := d.client.ListEvents(ctx,
		config.SourceID.ValueString(),
		config.Type.ValueString(),
		config.DateFrom.ValueString(),
		config.DateTo.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing events", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(events))
	for _, e := range events {
		obj, diags := types.ObjectValue(eventAttrTypes, map[string]attr.Value{
			"id":            types.StringValue(e.ID),
			"source_id":     types.StringValue(e.Source.ID),
			"type":          types.StringValue(e.Type),
			"text":          types.StringValue(e.Text),
			"time":          types.StringValue(e.Time),
			"creation_time": types.StringValue(e.CreationTime),
			"last_updated":  types.StringValue(e.LastUpdated),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: eventAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Events = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
