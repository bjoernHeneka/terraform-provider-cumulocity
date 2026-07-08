package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &loginOptionRawResource{}
var _ resource.ResourceWithImportState = &loginOptionRawResource{}

type loginOptionRawResource struct {
	client *client.Client
}

func NewLoginOptionRawResource() resource.Resource {
	return &loginOptionRawResource{}
}

type loginOptionRawModel struct {
	ID         types.String `tfsdk:"id"`
	Body       types.String `tfsdk:"body"`
	ConfigJSON types.String `tfsdk:"config_json"`
}

func (r *loginOptionRawResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login_option_raw"
}

func (r *loginOptionRawResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity login option (authentication configuration) via a " +
			"verbatim JSON body. Use this instead of cumulocity_login_option when you need to " +
			"manage the full authConfig, including nested/type-specific fields such as " +
			"tokenRequest, authorizationRequest, onNewUser.dynamicMapping or " +
			"signatureVerificationConfig (e.g. OAUTH2 / SSO providers). " +
			"Corresponds to POST/GET/PUT/DELETE /tenant/loginOptions. " +
			"Requires ROLE_TENANT_ADMIN or ROLE_TENANT_MANAGEMENT_ADMIN.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the login option assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"body": schema.StringAttribute{
				Required: true,
				Description: "The complete authConfig as a JSON string, sent verbatim to the API. " +
					"Use jsonencode({...}) to build it from an HCL object so formatting stays stable. " +
					"Note: this resource does not reconcile individual fields against the server — " +
					"drift detection is limited to whether the option still exists. The server masks " +
					"secrets (e.g. client_secret) in responses, so config_json is not a round-trippable " +
					"copy of body.",
			},
			"config_json": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				Description: "The complete raw JSON payload as returned by the API after the last " +
					"create/update/read, including server-added and normalized fields. Parse with " +
					"jsondecode(). Marked sensitive as it may echo request templates.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *loginOptionRawResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *loginOptionRawResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loginOptionRawModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateLoginOptionRaw(ctx, []byte(plan.Body.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error creating login option", err.Error())
		return
	}

	plan.ID = types.StringValue(result.ID)
	plan.ConfigJSON = types.StringValue(string(result.Raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *loginOptionRawResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loginOptionRawModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetLoginOption(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading login option", err.Error())
		return
	}

	// Only refresh the observed server payload. body is intentionally left
	// untouched: the server enriches and masks the config, so reconciling body
	// from the response would cause perpetual diffs.
	state.ConfigJSON = types.StringValue(string(result.Raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *loginOptionRawResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loginOptionRawModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateLoginOptionRaw(ctx, plan.ID.ValueString(), []byte(plan.Body.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error updating login option", err.Error())
		return
	}

	plan.ConfigJSON = types.StringValue(string(result.Raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *loginOptionRawResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loginOptionRawModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLoginOption(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting login option", err.Error())
	}
}

// ImportState imports by ID (or type). body is not populated by import — after
// importing, define body in your configuration to match the desired state.
// You can use the config_json output as a starting point, replacing any masked
// secret placeholders (e.g. "****") with real values.
func (r *loginOptionRawResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
