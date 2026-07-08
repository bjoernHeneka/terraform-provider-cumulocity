# Pending operations queued for a specific device
data "cumulocity_operations" "pending" {
  device_id = cumulocity_managed_object.device.id
  status    = "PENDING"
}

output "pending_operation_ids" {
  value = [for op in data.cumulocity_operations.pending.operations : op.id]
}
