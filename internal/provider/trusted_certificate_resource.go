package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/org-codebee/terraform-provider-cumulocity/internal/client"
)

var _ resource.Resource = &trustedCertificateResource{}
var _ resource.ResourceWithImportState = &trustedCertificateResource{}

type trustedCertificateResource struct {
	client *client.Client
}

func NewTrustedCertificateResource() resource.Resource {
	return &trustedCertificateResource{}
}

type trustedCertificateModel struct {
	ID                      types.String `tfsdk:"id"`
	TenantID                types.String `tfsdk:"tenant_id"`
	Name                    types.String `tfsdk:"name"`
	CertInPemFormat         types.String `tfsdk:"cert_in_pem_format"`
	Status                  types.String `tfsdk:"status"`
	AutoRegistrationEnabled types.Bool   `tfsdk:"auto_registration_enabled"`
	Fingerprint             types.String `tfsdk:"fingerprint"`
	AlgorithmName           types.String `tfsdk:"algorithm_name"`
	Issuer                  types.String `tfsdk:"issuer"`
	NotAfter                types.String `tfsdk:"not_after"`
	NotBefore               types.String `tfsdk:"not_before"`
	Self                    types.String `tfsdk:"self"`
}

func (r *trustedCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trusted_certificate"
}

func (r *trustedCertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Uploads and manages a trusted X.509 certificate for a Cumulocity tenant. " +
			"Devices use these certificates to establish connections with the platform. " +
			"Corresponds to POST/GET/PUT/DELETE /tenant/tenants/{tenantId}/trusted-certificates/{fingerprint}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {tenantId}/{fingerprint}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The tenant ID to upload the certificate to. Defaults to the provider's tenant_id. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert_in_pem_format": schema.StringAttribute{
				Required:    true,
				Description: "The trusted certificate in PEM format. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Required:    true,
				Description: "Whether the certificate is active: ENABLED or DISABLED.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable name for the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_registration_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether devices can auto-register using this certificate.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "Unique fingerprint of the certificate (assigned by Cumulocity).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"algorithm_name": schema.StringAttribute{
				Computed:    true,
				Description: "Algorithm used to encode the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issuer": schema.StringAttribute{
				Computed:    true,
				Description: "The organization that signed the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"not_after": schema.StringAttribute{
				Computed:    true,
				Description: "End of the certificate's validity period.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"not_before": schema.StringAttribute{
				Computed:    true,
				Description: "Start of the certificate's validity period.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the trusted certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *trustedCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trustedCertificateResource) effectiveTenantID(configured string) string {
	if configured != "" {
		return configured
	}
	return r.client.TenantID
}

func (r *trustedCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trustedCertificateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.effectiveTenantID(plan.TenantID.ValueString())
	if tenantID == "" {
		resp.Diagnostics.AddError("Missing tenant_id", "Set tenant_id on the resource or tenant_id on the provider.")
		return
	}

	cert := client.TrustedCertificate{
		CertInPemFormat:         plan.CertInPemFormat.ValueString(),
		Status:                  plan.Status.ValueString(),
		Name:                    plan.Name.ValueString(),
		AutoRegistrationEnabled: plan.AutoRegistrationEnabled.ValueBool(),
	}

	result, err := r.client.AddTrustedCertificate(ctx, tenantID, cert)
	if err != nil {
		resp.Diagnostics.AddError("Error adding trusted certificate", err.Error())
		return
	}

	r.apiToState(result, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *trustedCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trustedCertificateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.effectiveTenantID(state.TenantID.ValueString())
	result, err := r.client.GetTrustedCertificate(ctx, tenantID, state.Fingerprint.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading trusted certificate", err.Error())
		return
	}

	// Preserve PEM from state — API may not return it on GET.
	result.CertInPemFormat = state.CertInPemFormat.ValueString()
	r.apiToState(result, tenantID, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *trustedCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan trustedCertificateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.effectiveTenantID(plan.TenantID.ValueString())
	updateReq := client.TrustedCertificateUpdateRequest{
		Name:                    plan.Name.ValueString(),
		Status:                  plan.Status.ValueString(),
		AutoRegistrationEnabled: plan.AutoRegistrationEnabled.ValueBool(),
	}

	result, err := r.client.UpdateTrustedCertificate(ctx, tenantID, plan.Fingerprint.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating trusted certificate", err.Error())
		return
	}

	result.CertInPemFormat = plan.CertInPemFormat.ValueString()
	r.apiToState(result, tenantID, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *trustedCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trustedCertificateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.effectiveTenantID(state.TenantID.ValueString())
	if err := r.client.DeleteTrustedCertificate(ctx, tenantID, state.Fingerprint.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting trusted certificate", err.Error())
	}
}

// ImportState supports "{tenantId}/{fingerprint}".
func (r *trustedCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idx := len(req.ID) - 1
	for idx >= 0 && req.ID[idx] != '/' {
		idx--
	}
	if idx <= 0 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{tenantId}/{fingerprint}'.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), req.ID[:idx])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fingerprint"), req.ID[idx+1:])...)
}

func (r *trustedCertificateResource) apiToState(cert *client.TrustedCertificate, tenantID string, m *trustedCertificateModel) {
	m.TenantID = types.StringValue(tenantID)
	m.Fingerprint = types.StringValue(cert.Fingerprint)
	m.ID = types.StringValue(tenantID + "/" + cert.Fingerprint)
	m.Name = types.StringValue(cert.Name)
	m.CertInPemFormat = types.StringValue(cert.CertInPemFormat)
	m.Status = types.StringValue(cert.Status)
	m.AutoRegistrationEnabled = types.BoolValue(cert.AutoRegistrationEnabled)
	m.AlgorithmName = types.StringValue(cert.AlgorithmName)
	m.Issuer = types.StringValue(cert.Issuer)
	m.NotAfter = types.StringValue(cert.NotAfter)
	m.NotBefore = types.StringValue(cert.NotBefore)
	m.Self = types.StringValue(cert.Self)
}
