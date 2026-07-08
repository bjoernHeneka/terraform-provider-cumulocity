resource "cumulocity_managed_object" "fleet" {
  name            = "Field Devices"
  is_device_group = true
}

# Send a restart operation to every device in the group, spacing out the
# individual operations by 5 seconds each.
resource "cumulocity_bulk_operation" "restart_fleet" {
  group_id      = cumulocity_managed_object.fleet.id
  start_date    = "2025-01-01T12:00:00Z"
  creation_ramp = 5

  operation_prototype_json = jsonencode({
    c8y_Restart = {}
  })
}

output "bulk_operation_status" {
  value = cumulocity_bulk_operation.restart_fleet.general_status
}
