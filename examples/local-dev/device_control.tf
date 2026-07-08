
# ── Device Registration ───────────────────────────────────────────────────────

# Register a device and immediately accept it.
resource "cumulocity_new_device_request" "edge_gateway" {
  device_id = "edge-gateway-001"
  status    = "ACCEPTED"
}

# Auto-generate device credentials after registration.
resource "cumulocity_device_credentials" "edge_gateway" {
  device_id = cumulocity_new_device_request.edge_gateway.device_id
}

output "edge_gateway_username" {
  value = cumulocity_device_credentials.edge_gateway.username
}

output "edge_gateway_password" {
  value     = cumulocity_device_credentials.edge_gateway.password
  sensitive = true
}

# ── Device Operations ─────────────────────────────────────────────────────────

# Send a restart command to a device.
# resource "cumulocity_device_operation" "restart" {
#   device_id      = cumulocity_managed_object.test_device.id
#   description    = "Restart device via Terraform"
#   fragments_json = jsonencode({ c8y_Restart = {} })
# }

# Send a shell command to a device.
# resource "cumulocity_device_operation" "shell_cmd" {
#   device_id      = cumulocity_managed_object.test_device.id
#   description    = "List running processes"
#   fragments_json = jsonencode({ c8y_Command = { text = "ps aux" } })
# }

# ── Bulk Operations ────────────────────────────────────────────────────────────

# Schedule a restart for all devices in the test group.
# resource "cumulocity_bulk_operation" "group_restart" {
#   group_id                = cumulocity_managed_object.test_group.id
#   start_date              = "2025-06-01T08:00:00Z"
#   creation_ramp           = 5.0
#   operation_prototype_json = jsonencode({ c8y_Restart = {} })
# }

