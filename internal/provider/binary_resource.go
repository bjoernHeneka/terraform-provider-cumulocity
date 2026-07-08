package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &binaryResource{}
var _ resource.ResourceWithImportState = &binaryResource{}

type binaryResource struct {
	client *client.Client
}

func NewBinaryResource() resource.Resource {
	return &binaryResource{}
}

type binaryModel struct {
	ID          types.String `tfsdk:"id"`
	File        types.String `tfsdk:"file"`
	FileHash    types.String `tfsdk:"file_hash"`
	Name        types.String `tfsdk:"name"`
	ContentType types.String `tfsdk:"content_type"`
	Length      types.Int64  `tfsdk:"length"`
	Owner       types.String `tfsdk:"owner"`
	Self        types.String `tfsdk:"self"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *binaryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_binary"
}

func (r *binaryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads a file to the Cumulocity inventory binary store (/inventory/binaries). " +
			"Changing file, file_hash, name, or content_type forces the creation of a new binary resource (the old one is deleted). " +
			"Corresponds to POST /inventory/binaries (create), GET /inventory/managedObjects/{id} (read), " +
			"and DELETE /inventory/binaries/{id} (delete).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Binary managed object ID assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file": schema.StringAttribute{
				Required:    true,
				Description: "Local filesystem path to the file to upload. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Description: "Hash of the file content (e.g. filemd5(\"path/to/file\")). When this value changes, the binary is re-uploaded. Use this to trigger a re-upload when the file path stays the same but the content changes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name for the binary managed object. Defaults to the base filename. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "MIME content type of the file, e.g. `application/zip`. Defaults to `application/octet-stream`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"length": schema.Int64Attribute{
				Computed:    true,
				Description: "File size in bytes as reported by the platform.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Username of the owner of the binary managed object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the binary managed object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of the last update.",
			},
		},
	}
}

func (r *binaryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *binaryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan binaryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	binary, err := r.client.UploadBinary(ctx,
		plan.File.ValueString(),
		plan.Name.ValueString(),
		plan.ContentType.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error uploading binary", err.Error())
		return
	}

	r.apiToState(binary, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *binaryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state binaryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	binary, err := r.client.GetBinary(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading binary", err.Error())
		return
	}

	r.apiToState(binary, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all meaningful attributes are RequiresReplace.
func (r *binaryResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *binaryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state binaryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBinary(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting binary", err.Error())
	}
}

func (r *binaryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *binaryResource) apiToState(b *client.Binary, m *binaryModel) {
	m.ID = types.StringValue(b.ID)
	m.Name = types.StringValue(b.Name)
	m.ContentType = types.StringValue(b.ContentType)
	m.Length = types.Int64Value(b.Length)
	m.Owner = types.StringValue(b.Owner)
	m.Self = types.StringValue(b.Self)
	m.LastUpdated = types.StringValue(b.LastUpdated)
}
