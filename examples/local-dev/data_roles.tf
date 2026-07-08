
# Einzelne Rolle nachschlagen
data "cumulocity_role" "alarm_admin" {
  name = "ROLE_ALARM_ADMIN"
}

# Alle verfügbaren Rollen auflisten
data "cumulocity_roles" "all" {}

# Nur Alarm-Rollen (Filter)
data "cumulocity_roles" "alarm_roles" {
  name_filter = "ALARM"
}

# Outputs zur Überprüfung
output "alarm_admin_role_self" {
  value = data.cumulocity_role.alarm_admin.self
}

output "all_role_names" {
  value = [for r in data.cumulocity_roles.all.roles : r.name]
}

output "alarm_role_names" {
  value = [for r in data.cumulocity_roles.alarm_roles.roles : r.name]
}
