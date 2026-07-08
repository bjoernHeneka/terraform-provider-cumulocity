# List all global roles
data "cumulocity_roles" "all" {}

# Narrow to roles whose name contains "ALARM" (case-insensitive)
data "cumulocity_roles" "alarm_roles" {
  name_filter = "ALARM"
}

output "all_role_names" {
  value = [for r in data.cumulocity_roles.all.roles : r.name]
}

output "alarm_role_names" {
  value = [for r in data.cumulocity_roles.alarm_roles.roles : r.name]
}
