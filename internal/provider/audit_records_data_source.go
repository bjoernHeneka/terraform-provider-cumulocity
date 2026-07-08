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

var _ datasource.DataSource = &auditRecordsDataSource{}

type auditRecordsDataSource struct {
	client *client.Client
}

func NewAuditRecordsDataSource() datasource.DataSource {
	return &auditRecordsDataSource{}
}

type auditRecordsDataModel struct {
	SourceID     types.String `tfsdk:"source_id"`
	Type         types.String `tfsdk:"type"`
	User         types.String `tfsdk:"user"`
	Application  types.String `tfsdk:"application"`
	AuditRecords types.List   `tfsdk:"audit_records"`
}

var auditRecordAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"source_id":     types.StringType,
	"activity":      types.StringType,
	"text":          types.StringType,
	"time":          types.StringType,
	"type":          types.StringType,
	"user":          types.StringType,
	"application":   types.StringType,
	"severity":      types.StringType,
	"creation_time": types.StringType,
}

func (d *auditRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_records"
}

func (d *auditRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of audit records, optionally filtered by source, type, user, and/or application. " +
			"All pages are followed automatically. " +
			"Corresponds to GET /audit/auditRecords.",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter audit records by the platform component or managed object ID.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by audit record type, e.g. `Operation`, `User`, `Alarm`.",
			},
			"user": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the username who carried out the activity.",
			},
			"application": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by the application name from which the audit was carried out.",
			},
			"audit_records": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching audit records.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, Description: "Audit record ID."},
						"source_id":     schema.StringAttribute{Computed: true, Description: "Source platform component or managed object ID."},
						"activity":      schema.StringAttribute{Computed: true, Description: "Summary of the action."},
						"text":          schema.StringAttribute{Computed: true, Description: "Detailed description."},
						"time":          schema.StringAttribute{Computed: true, Description: "ISO 8601 time of the audit event."},
						"type":          schema.StringAttribute{Computed: true, Description: "Platform component type."},
						"user":          schema.StringAttribute{Computed: true, Description: "User who performed the action."},
						"application":   schema.StringAttribute{Computed: true, Description: "Application that performed the action."},
						"severity":      schema.StringAttribute{Computed: true, Description: "Severity level."},
						"creation_time": schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
					},
				},
			},
		},
	}
}

func (d *auditRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *auditRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config auditRecordsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	records, err := d.client.ListAuditRecords(ctx,
		config.SourceID.ValueString(),
		config.Type.ValueString(),
		config.User.ValueString(),
		config.Application.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error listing audit records", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(records))
	for _, a := range records {
		obj, diags := types.ObjectValue(auditRecordAttrTypes, map[string]attr.Value{
			"id":            types.StringValue(a.ID),
			"source_id":     types.StringValue(a.Source.ID),
			"activity":      types.StringValue(a.Activity),
			"text":          types.StringValue(a.Text),
			"time":          types.StringValue(a.Time),
			"type":          types.StringValue(a.Type),
			"user":          types.StringValue(a.User),
			"application":   types.StringValue(a.Application),
			"severity":      types.StringValue(a.Severity),
			"creation_time": types.StringValue(a.CreationTime),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: auditRecordAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.AuditRecords = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
