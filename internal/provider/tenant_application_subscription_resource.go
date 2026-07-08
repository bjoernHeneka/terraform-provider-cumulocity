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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &tenantApplicationSubscriptionResource{}
var _ resource.ResourceWithImportState = &tenantApplicationSubscriptionResource{}

type tenantApplicationSubscriptionResource struct {
	client *client.Client
}

func NewTenantApplicationSubscriptionResource() resource.Resource {
	return &tenantApplicationSubscriptionResource{}
}

type tenantApplicationSubscriptionModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *tenantApplicationSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_application_subscription"
}

func (r *tenantApplicationSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Subscribes a tenant to an application. " +
			"Corresponds to POST/DELETE /tenant/tenants/{tenantId}/applications. " +
			"Requires ROLE_APPLICATION_MANAGEMENT_ADMIN or ROLE_TENANT_MANAGEMENT_ADMIN/UPDATE.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite ID in the format 'tenantId/applicationId'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Required:    true,
				Description: "The tenant ID to subscribe (e.g. 't0071234'). Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_id": schema.StringAttribute{
				Required:    true,
				Description: "The application ID to subscribe to. Immutable — changing forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tenantApplicationSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tenantApplicationSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantApplicationSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := plan.TenantID.ValueString()
	applicationID := plan.ApplicationID.ValueString()

	_, err := r.client.SubscribeApplication(ctx, tenantID, applicationID)
	if err != nil {
		resp.Diagnostics.AddError("Error subscribing to application", err.Error())
		return
	}

	// Set ALL fields explicitly (including input fields from plan)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantID, applicationID))
	plan.TenantID = types.StringValue(tenantID)
	plan.ApplicationID = types.StringValue(applicationID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *tenantApplicationSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantApplicationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := state.TenantID.ValueString()
	applicationID := state.ApplicationID.ValueString()

	_, err := r.client.GetSubscribedApplication(ctx, tenantID, applicationID)
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading application subscription",
			fmt.Sprintf("Could not read subscription for tenant %q and application %q: %s", tenantID, applicationID, err.Error()),
		)
		return
	}

	// IMPORTANT: Always set ALL fields in Read, including ID
	// Keep the same ID format as in Create
	state.ID = types.StringValue(fmt.Sprintf("%s/%s", tenantID, applicationID))
	state.TenantID = types.StringValue(tenantID)
	state.ApplicationID = types.StringValue(applicationID)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *tenantApplicationSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource has no updatable fields - tenant_id and application_id are immutable
	// If either changes, it forces a new resource (RequiresReplace)
	resp.Diagnostics.AddError(
		"Update not supported",
		"Tenant application subscriptions cannot be updated. Changes to tenant_id or application_id will force a new resource.",
	)
}

func (r *tenantApplicationSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantApplicationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := state.TenantID.ValueString()
	applicationID := state.ApplicationID.ValueString()

	if err := r.client.UnsubscribeApplication(ctx, tenantID, applicationID); err != nil {
		resp.Diagnostics.AddError("Error unsubscribing from application", err.Error())
	}
}

func (r *tenantApplicationSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import expects format: tenantId/applicationId
	id := req.ID
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected format: tenantId/applicationId (e.g. t0071234/12345), got: %q", id),
		)
		return
	}

	tenantID := parts[0]
	applicationID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_id"), applicationID)...)
}
