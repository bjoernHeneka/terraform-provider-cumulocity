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

var _ datasource.DataSource = &operationsDataSource{}

type operationsDataSource struct {
	client *client.Client
}

func NewOperationsDataSource() datasource.DataSource {
	return &operationsDataSource{}
}

type operationsDataModel struct {
	DeviceID   types.String `tfsdk:"device_id"`
	Status     types.String `tfsdk:"status"`
	Operations types.List   `tfsdk:"operations"`
}

var operationAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"device_id":         types.StringType,
	"status":            types.StringType,
	"failure_reason":    types.StringType,
	"creation_time":     types.StringType,
	"bulk_operation_id": types.Int64Type,
	"self":              types.StringType,
}

func (d *operationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operations"
}

func (d *operationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of device operations, optionally filtered by device ID and/or status. " +
			"All pages are followed automatically. " +
			"Corresponds to GET /devicecontrol/operations.",
		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter operations by device ID.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by status: PENDING, EXECUTING, SUCCESSFUL, or FAILED.",
			},
			"operations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching operations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true, Description: "Operation ID."},
						"device_id":         schema.StringAttribute{Computed: true, Description: "Target device ID."},
						"status":            schema.StringAttribute{Computed: true, Description: "Operation status."},
						"failure_reason":    schema.StringAttribute{Computed: true, Description: "Failure reason, if any."},
						"creation_time":     schema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"bulk_operation_id": schema.Int64Attribute{Computed: true, Description: "Parent bulk operation ID."},
						"self":              schema.StringAttribute{Computed: true, Description: "Self-link URL."},
					},
				},
			},
		},
	}
}

func (d *operationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *operationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config operationsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ops, err := d.client.ListOperations(ctx, config.DeviceID.ValueString(), config.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing operations", err.Error())
		return
	}

	elems := make([]attr.Value, 0, len(ops))
	for _, op := range ops {
		obj, diags := types.ObjectValue(operationAttrTypes, map[string]attr.Value{
			"id":                types.StringValue(op.ID),
			"device_id":         types.StringValue(op.DeviceID),
			"status":            types.StringValue(op.Status),
			"failure_reason":    types.StringValue(op.FailureReason),
			"creation_time":     types.StringValue(op.CreationTime),
			"bulk_operation_id": types.Int64Value(op.BulkOperationID),
			"self":              types.StringValue(op.Self),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: operationAttrTypes}, elems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Operations = list
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
