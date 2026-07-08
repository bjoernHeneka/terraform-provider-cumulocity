package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &applicationBinaryResource{}
var _ resource.ResourceWithImportState = &applicationBinaryResource{}

type applicationBinaryResource struct {
	client *client.Client
}

func NewApplicationBinaryResource() resource.Resource {
	return &applicationBinaryResource{}
}

type applicationBinaryModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	File          types.String `tfsdk:"file"`
	FileHash      types.String `tfsdk:"file_hash"`
	BinaryID      types.String `tfsdk:"binary_id"`
	Name          types.String `tfsdk:"name"`
	Length        types.Int64  `tfsdk:"length"`
	Created       types.String `tfsdk:"created"`
	DownloadURL   types.String `tfsdk:"download_url"`
}

func (r *applicationBinaryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_binary"
}

func (r *applicationBinaryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads a ZIP archive to a Cumulocity application. " +
			"Each upload creates a new binary version and sets it as the active version. " +
			"All attributes are immutable — changing file or file_hash triggers a new upload (old binary is deleted). " +
			"Corresponds to POST /application/applications/{id}/binaries.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {applicationId}/{binaryId}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the application to upload the binary to. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file": schema.StringAttribute{
				Required:    true,
				Description: "Local filesystem path to the ZIP file to upload. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_hash": schema.StringAttribute{
				Optional:    true,
				Description: "Hash of the file content, e.g. filemd5(\"path/to/app.zip\"). When this value changes, the binary is re-uploaded. Use this to trigger a re-upload when the file path stays the same but the content changes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"binary_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the uploaded binary attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Filename of the uploaded archive.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"length": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the uploaded archive in bytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp when the binary was uploaded.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"download_url": schema.StringAttribute{
				Computed:    true,
				Description: "Download URL for the uploaded archive.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *applicationBinaryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *applicationBinaryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationBinaryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := plan.ApplicationID.ValueString()
	app, err := r.client.UploadApplicationBinary(ctx, appID, plan.File.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error uploading application binary", err.Error())
		return
	}

	// The upload response contains activeVersionId — that is the binary ID.
	binaryID := app.ActiveVersionID
	if binaryID == "" {
		// Fallback: list binaries and take the most recent one by name.
		binaries, listErr := r.client.GetApplicationBinaries(ctx, appID)
		if listErr != nil {
			resp.Diagnostics.AddError("Error listing binaries after upload", listErr.Error())
			return
		}
		if len(binaries) == 0 {
			resp.Diagnostics.AddError("No binaries found after upload", "The upload appeared to succeed but no binary was found.")
			return
		}
		binaryID = binaries[len(binaries)-1].ID
	}

	// Fetch the binary metadata for computed attributes.
	binary, err := r.client.GetApplicationBinaryByID(ctx, appID, binaryID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading binary metadata after upload", err.Error())
		return
	}

	r.apiToState(binary, appID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationBinaryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationBinaryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ApplicationID.ValueString()
	binaryID := state.BinaryID.ValueString()

	binary, err := r.client.GetApplicationBinaryByID(ctx, appID, binaryID)
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading application binary", err.Error())
		return
	}

	r.apiToState(binary, appID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all meaningful attributes are RequiresReplace.
func (r *applicationBinaryResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *applicationBinaryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationBinaryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplicationBinary(ctx, state.ApplicationID.ValueString(), state.BinaryID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting application binary", err.Error())
	}
}

// ImportState supports "{applicationId}/{binaryId}".
func (r *applicationBinaryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{applicationId}/{binaryId}'.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("binary_id"), parts[1])...)
}

func (r *applicationBinaryResource) apiToState(b *client.ApplicationBinary, appID string, m *applicationBinaryModel) {
	m.ApplicationID = types.StringValue(appID)
	m.BinaryID = types.StringValue(b.ID)
	m.Name = types.StringValue(b.Name)
	m.Length = types.Int64Value(b.Length)
	m.Created = types.StringValue(b.Created)
	m.DownloadURL = types.StringValue(b.DownloadURL)
	m.ID = types.StringValue(appID + "/" + b.ID)
}
