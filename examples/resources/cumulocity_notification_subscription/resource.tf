# Subscribe to alarms and events for a specific device (managed object context).
resource "cumulocity_notification_subscription" "device_alerts" {
  context      = "mo"
  source_id    = cumulocity_managed_object.gateway.id
  subscription = "GatewayAlerts"
  apis         = ["alarms", "events"]
  type_filter  = "'c8y_UnavailabilityAlarm' or 'c8y_ConnectivityAlarm'"
}

# Subscribe to all inventory changes at the tenant level.
resource "cumulocity_notification_subscription" "tenant_inventory" {
  context      = "tenant"
  subscription = "TenantInventory"
  apis         = ["managedobjects"]
}

# Subscribe to all data types for a device, forwarding only specific fragments.
resource "cumulocity_notification_subscription" "device_full" {
  context           = "mo"
  source_id         = cumulocity_managed_object.sensor.id
  subscription      = "SensorFull"
  apis              = ["*"]
  fragments_to_copy = ["c8y_Temperature", "c8y_Position"]
  non_persistent    = false
}
