
# Nachschlagen per Name
data "cumulocity_inventory_role" "restart" {
  name = "Operations: Restart Device"
}

# Nachschlagen per numerischer ID
# data "cumulocity_inventory_role" "by_id" {
#   id = 4
# }

# Alle Inventory Roles auflisten
data "cumulocity_inventory_roles" "all" {}

# Gefiltert — nur Operations-Rollen
data "cumulocity_inventory_roles" "operations" {
  name_filter = "Operations"
}

output "restart_role_id" {
  value = data.cumulocity_inventory_role.restart.id
}

output "restart_role_permissions" {
  value = data.cumulocity_inventory_role.restart.permissions
}

output "all_inventory_role_names" {
  value = [for r in data.cumulocity_inventory_roles.all.roles : r.name]
}
