---
page_title: "Resource: cumulocity_notification_subscription"
description: |-
  Manages a Cumulocity Notification 2.0 subscription.
---

# cumulocity_notification_subscription

Manages a Cumulocity Notification 2.0 subscription. A subscription defines which data (APIs, type filters) is forwarded for a given device or tenant context. Subscriptions are immutable — all attributes trigger a resource replacement when changed.

Corresponds to `POST/GET/DELETE /notification2/subscriptions/{id}`.

## Example Usage

### Device (managed object) subscription

```hcl
# Subscribe to alarms and events for a specific device (managed object context).
resource "cumulocity_notification_subscription" "device_alerts" {
  context      = "mo"
  source_id    = cumulocity_managed_object.gateway.id
  subscription = "GatewayAlerts"
  apis         = ["alarms", "events"]
  type_filter  = "'c8y_UnavailabilityAlarm' or 'c8y_ConnectivityAlarm'"
}
```

### Tenant subscription

```hcl
# Subscribe to all inventory changes at the tenant level.
resource "cumulocity_notification_subscription" "tenant_inventory" {
  context      = "tenant"
  subscription = "TenantInventory"
  apis         = ["managedobjects"]
}
```

### Forwarding selected fragments

```hcl
# Subscribe to all data types for a device, forwarding only specific fragments.
resource "cumulocity_notification_subscription" "device_full" {
  context           = "mo"
  source_id         = cumulocity_managed_object.sensor.id
  subscription      = "SensorFull"
  apis              = ["*"]
  fragments_to_copy = ["c8y_Temperature", "c8y_Position"]
  non_persistent    = false
}
```

## Schema

### Required

- `context` (String) — The context within which the subscription is processed. Must be `mo` (managed object) or `tenant`. When set to `mo`, `source_id` is required. Changing this forces a new resource.
- `subscription` (String) — The subscription name, unique within its context. Only alphanumeric characters are allowed. Changing this forces a new resource.

### Optional

- `source_id` (String) — The managed object ID to associate with the subscription. Required when `context` is `mo`. Changing this forces a new resource.
- `apis` (List of String) — List of APIs to subscribe to. Valid values: `alarms`, `alarmsWithChildren`, `events`, `eventsWithChildren`, `managedobjects`, `measurements`, `operations`, `*`. Changing this forces a new resource.
- `type_filter` (String) — OData type filter expression, e.g. `'c8y_Speed' or 'c8y_LocationUpdate'`. Changing this forces a new resource.
- `fragments_to_copy` (List of String) — List of custom fragment names to include in forwarded data. If empty, data is forwarded as-is. Changing this forces a new resource.
- `non_persistent` (Boolean) — When `true`, messages may be lost if no consumer is connected. Defaults to `false`. Changing this forces a new resource.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `self` (String) — Self-link URL of the subscription.

## Import

Import an existing subscription by its ID:

```shell
terraform import cumulocity_notification_subscription.device_alerts 20200301
```
