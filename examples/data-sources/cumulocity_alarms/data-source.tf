# All active alarms for a specific device
data "cumulocity_alarms" "active" {
  source_id = cumulocity_managed_object.device.id
  status    = "ACTIVE"
}

# All critical alarms across the tenant
data "cumulocity_alarms" "critical" {
  severity = "CRITICAL"
}

output "active_alarm_count" {
  value = length(data.cumulocity_alarms.active.alarms)
}

output "critical_alarm_texts" {
  value = [for a in data.cumulocity_alarms.critical.alarms : a.text]
}
