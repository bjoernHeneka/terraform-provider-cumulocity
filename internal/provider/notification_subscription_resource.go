package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/bjoernHeneka/terraform-provider-cumulocity/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &notificationSubscriptionResource{}
var _ resource.ResourceWithImportState = &notificationSubscriptionResource{}

type notificationSubscriptionResource struct {
	client *client.Client
}

func NewNotificationSubscriptionResource() resource.Resource {
	return &notificationSubscriptionResource{}
}

type notificationSubscriptionModel struct {
	ID              types.String `tfsdk:"id"`
	Context         types.String `tfsdk:"context"`
	Subscription    types.String `tfsdk:"subscription"`
	SourceID        types.String `tfsdk:"source_id"`
	APIs            types.List   `tfsdk:"apis"`
	TypeFilter      types.String `tfsdk:"type_filter"`
	FragmentsToCopy types.List   `tfsdk:"fragments_to_copy"`
	NonPersistent   types.Bool   `tfsdk:"non_persistent"`
	Self            types.String `tfsdk:"self"`
}

func (r *notificationSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_subscription"
}

func (r *notificationSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cumulocity Notification 2.0 subscription. A subscription defines which data " +
			"(APIs, type filters) is forwarded for a given device or tenant context. " +
			"Subscriptions are immutable — all attributes trigger a resource replacement when changed. " +
			"Corresponds to POST/GET/DELETE /notification2/subscriptions/{id}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier assigned by Cumulocity.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"context": schema.StringAttribute{
				Required:    true,
				Description: "The context within which the subscription is processed. Must be `mo` (managed object) or `tenant`. When set to `mo`, `source_id` is required. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subscription": schema.StringAttribute{
				Required:    true,
				Description: "The subscription name, unique within its context. Only alphanumeric characters are allowed. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_id": schema.StringAttribute{
				Optional:    true,
				Description: "The managed object ID to associate with the subscription. Required when `context` is `mo`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"apis": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of APIs to subscribe to. Valid values: `alarms`, `alarmsWithChildren`, `events`, `eventsWithChildren`, `managedobjects`, `measurements`, `operations`, `*`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"type_filter": schema.StringAttribute{
				Optional:    true,
				Description: "OData type filter expression, e.g. `'c8y_Speed' or 'c8y_LocationUpdate'`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fragments_to_copy": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of custom fragment names to include in forwarded data. If empty, data is forwarded as-is. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"non_persistent": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When `true`, messages may be lost if no consumer is connected. Defaults to `false`. Changing this value forces a new resource.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"self": schema.StringAttribute{
				Computed:    true,
				Description: "Self-link URL of the subscription.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *notificationSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := r.modelToRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.CreateNotificationSubscription(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification subscription", err.Error())
		return
	}

	resp.Diagnostics.Append(r.apiToState(ctx, sub, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.GetNotificationSubscription(ctx, state.ID.ValueString())
	if errors.Is(err, client.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification subscription", err.Error())
		return
	}

	resp.Diagnostics.Append(r.apiToState(ctx, sub, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// Update is not implemented — all attributes are RequiresReplace.
func (r *notificationSubscriptionResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *notificationSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNotificationSubscription(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting notification subscription", err.Error())
	}
}

func (r *notificationSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sub, err := r.client.GetNotificationSubscription(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing notification subscription", err.Error())
		return
	}
	var state notificationSubscriptionModel
	resp.Diagnostics.Append(r.apiToState(ctx, sub, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *notificationSubscriptionResource) modelToRequest(ctx context.Context, m notificationSubscriptionModel) (client.NotificationSubscriptionRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := client.NotificationSubscriptionRequest{
		Context:       m.Context.ValueString(),
		Subscription:  m.Subscription.ValueString(),
		NonPersistent: m.NonPersistent.ValueBool(),
	}

	if !m.SourceID.IsNull() && !m.SourceID.IsUnknown() {
		req.Source = &client.NotificationSubscriptionSource{ID: m.SourceID.ValueString()}
	}

	var apis []string
	if !m.APIs.IsNull() && !m.APIs.IsUnknown() {
		diags.Append(m.APIs.ElementsAs(ctx, &apis, false)...)
		if diags.HasError() {
			return req, diags
		}
	}

	var fragments []string
	if !m.FragmentsToCopy.IsNull() && !m.FragmentsToCopy.IsUnknown() {
		diags.Append(m.FragmentsToCopy.ElementsAs(ctx, &fragments, false)...)
		if diags.HasError() {
			return req, diags
		}
	}

	if len(apis) > 0 || (!m.TypeFilter.IsNull() && !m.TypeFilter.IsUnknown()) {
		filter := &client.NotificationSubscriptionFilter{}
		if len(apis) > 0 {
			filter.APIs = apis
		}
		if !m.TypeFilter.IsNull() && !m.TypeFilter.IsUnknown() {
			filter.TypeFilter = m.TypeFilter.ValueString()
		}
		req.SubscriptionFilter = filter
	}

	if len(fragments) > 0 {
		req.FragmentsToCopy = fragments
	}

	return req, diags
}

func (r *notificationSubscriptionResource) apiToState(ctx context.Context, sub *client.NotificationSubscription, m *notificationSubscriptionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(sub.ID)
	m.Self = types.StringValue(sub.Self)
	m.Context = types.StringValue(sub.Context)
	m.Subscription = types.StringValue(sub.Subscription)
	m.NonPersistent = types.BoolValue(sub.NonPersistent)

	if sub.Source != nil {
		m.SourceID = types.StringValue(sub.Source.ID)
	} else {
		m.SourceID = types.StringNull()
	}

	if sub.SubscriptionFilter != nil && len(sub.SubscriptionFilter.APIs) > 0 {
		elems := make([]attr.Value, len(sub.SubscriptionFilter.APIs))
		for i, a := range sub.SubscriptionFilter.APIs {
			elems[i] = types.StringValue(a)
		}
		list, d := types.ListValue(types.StringType, elems)
		diags.Append(d...)
		m.APIs = list
	} else {
		m.APIs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if sub.SubscriptionFilter != nil && sub.SubscriptionFilter.TypeFilter != "" {
		m.TypeFilter = types.StringValue(sub.SubscriptionFilter.TypeFilter)
	} else {
		m.TypeFilter = types.StringNull()
	}

	if len(sub.FragmentsToCopy) > 0 {
		elems := make([]attr.Value, len(sub.FragmentsToCopy))
		for i, f := range sub.FragmentsToCopy {
			elems[i] = types.StringValue(f)
		}
		list, d := types.ListValue(types.StringType, elems)
		diags.Append(d...)
		m.FragmentsToCopy = list
	} else {
		m.FragmentsToCopy = types.ListValueMust(types.StringType, []attr.Value{})
	}

	return diags
}
