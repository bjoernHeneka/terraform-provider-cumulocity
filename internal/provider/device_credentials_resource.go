package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &deviceCredentialsResource{}
var _ resource.ResourceWithImportState = &deviceCredentialsResource{}

type deviceCredentialsResource struct {
	client *client.Client
}

func NewDeviceCredentialsResource() resource.Resource {
	return &deviceCredentialsResource{}
}

type deviceCredentialsModel struct {
	ID            types.String `tfsdk:"id"`
	DeviceID      types.String `tfsdk:"device_id"`
	SecurityToken types.String `tfsdk:"security_token"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Self          types.String `tfsdk:"self"`
}

func (r *deviceCredentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_credentials"
}

func (r *deviceCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Requests auto-generated credentials for a device identified by its external ID. " +
			"Credentials are created once; username and password are returned only on creation and stored in state. " +
			"Destroy removes the resource from Terraform state but credentials persist in Cumulocity (they are tied to the device lifecycle). " +
			"Corresponds to POST /devicecontrol/deviceCredentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Terraform identifier (equals device_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.StringAttribute{
				Required:    true,
				Description: "External ID of the device (the device registration ID). Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"security_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "One-time security token submitted by the device during registration. Write-only — never read back from the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "Auto-generated username for the device.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Auto-generated password for the device. Only returned on creation and stored in Terraform state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Computed:    true,
				Description: "Tenant ID associated with these device credentials.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the device credentials.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *deviceCredentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceCredentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceCredentialsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	creds, err := r.client.CreateDeviceCredentials(ctx,
		plan.DeviceID.ValueString(),
		plan.SecurityToken.ValueString(),
	)
	if err != nil {
		detail := err.Error()
		if strings.Contains(detail, "403") {
			detail = fmt.Sprintf(
				"Access denied (HTTP 403).\n\n"+
					"POST /devicecontrol/deviceCredentials is restricted to the Cumulocity bootstrap user "+
					"(devicebootstrap). It cannot be called with regular admin credentials.\n\n"+
					"This endpoint is intended for the device bootstrap flow, not for admin automation. "+
					"Configure the provider with bootstrap credentials or remove this resource.\n\n"+
					"Original error: %s", err.Error())
		}
		resp.Diagnostics.AddError("Error creating device credentials", detail)
		return
	}

	plan.ID = types.StringValue(creds.ID)
	plan.DeviceID = types.StringValue(creds.ID)
	plan.Username = types.StringValue(creds.Username)
	plan.Password = types.StringValue(creds.Password)
	plan.TenantID = types.StringValue(creds.TenantID)
	plan.Self = types.StringValue(creds.Self)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read preserves existing state — the API has no GET endpoint for individual credentials.
func (r *deviceCredentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceCredentialsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all attributes are RequiresReplace.
func (r *deviceCredentialsResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

// Delete removes the resource from Terraform state.
// Cumulocity device credentials are tied to the device lifecycle and cannot be individually deleted via API.
func (r *deviceCredentialsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState supports import by device_id.
func (r *deviceCredentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), req.ID)...)
}
