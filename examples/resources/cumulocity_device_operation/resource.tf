resource "cumulocity_managed_object" "device" {
  name      = "Edge Gateway"
  type      = "c8y_Gateway"
  is_device = true
}

# Send a restart operation to the device. The operation starts in PENDING and
# is updated by the device as it is processed.
resource "cumulocity_device_operation" "restart" {
  device_id   = cumulocity_managed_object.device.id
  description = "Restart the gateway"
  fragments_json = jsonencode({
    c8y_Restart = {}
  })
}

# Send a shell command operation to the device.
resource "cumulocity_device_operation" "command" {
  device_id   = cumulocity_managed_object.device.id
  description = "List working directory"
  fragments_json = jsonencode({
    c8y_Command = {
      text = "ls -la"
    }
  })
}

output "restart_status" {
  value = cumulocity_device_operation.restart.status
}
