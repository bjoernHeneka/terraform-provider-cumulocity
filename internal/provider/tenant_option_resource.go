package provider

import (
	"context"
	"errors"
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

var _ resource.Resource = &tenantOptionResource{}
var _ resource.ResourceWithImportState = &tenantOptionResource{}

type tenantOptionResource struct {
	client *client.Client
}

func NewTenantOptionResource() resource.Resource {
	return &tenantOptionResource{}
}

type tenantOptionModel struct {
	ID       types.String `tfsdk:"id"`
	Category types.String `tfsdk:"category"`
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	Self     types.String `tfsdk:"self"`
}

func (r *tenantOptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_option"
}

func (r *tenantOptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates and manages a Cumulocity tenant option (category/key/value tuple). " +
			"Tenant options store per-tenant configuration consumed by Cumulocity and microservices. " +
			"Corresponds to POST/GET/PUT/DELETE /tenant/options/{category}/{key}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite Terraform identifier: {category}/{key}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"category": schema.StringAttribute{
				Required:    true,
				Description: "Category of the option, e.g. \"access.control\" or \"alarm.type.mapping\". Immutable — changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "Key of the option, unique within its category. Prefix with \"credentials.\" to store the value encrypted. Immutable — changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Value of the option. Marked sensitive because keys with a \"credentials.\" prefix are stored encrypted. Changing this value performs an in-place update.",
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the tenant option.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *tenantOptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tenantOptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt, err := r.client.CreateTenantOption(ctx, plan.Category.ValueString(), plan.Key.ValueString(), plan.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating tenant option", err.Error())
		return
	}

	r.apiToState(opt, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tenantOptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt, err := r.client.GetTenantOption(ctx, state.Category.ValueString(), state.Key.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading tenant option", err.Error())
		return
	}

	// Preserve the value from state — GET returns the raw (possibly encrypted) value
	// which may differ from what was set (e.g. credentials.* keys).
	opt.Value = state.Value.ValueString()
	r.apiToState(opt, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tenantOptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantOptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opt, err := r.client.UpdateTenantOption(ctx, plan.Category.ValueString(), plan.Key.ValueString(), plan.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating tenant option", err.Error())
		return
	}

	// Preserve the planned value (API may return masked value for credentials.* keys).
	opt.Value = plan.Value.ValueString()
	r.apiToState(opt, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tenantOptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantOptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTenantOption(ctx, state.Category.ValueString(), state.Key.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting tenant option", err.Error())
	}
}

// ImportState supports "{category}/{key}".
func (r *tenantOptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idx := strings.LastIndex(req.ID, "/")
	if idx <= 0 || idx == len(req.ID)-1 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected '{category}/{key}', e.g. 'access.control/allow.origin'.")
		return
	}
	category := req.ID[:idx]
	key := req.ID[idx+1:]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("category"), category)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), key)...)
}

func (r *tenantOptionResource) apiToState(opt *client.TenantOption, m *tenantOptionModel) {
	m.Category = types.StringValue(opt.Category)
	m.Key = types.StringValue(opt.Key)
	m.Value = types.StringValue(opt.Value)
	m.Self = types.StringValue(opt.Self)
	m.ID = types.StringValue(opt.Category + "/" + opt.Key)
}
