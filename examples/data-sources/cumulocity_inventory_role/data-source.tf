# Look up an inventory role by its exact name
data "cumulocity_inventory_role" "restart" {
  name = "Operations: Restart Device"
}

# Alternatively, look up by numeric ID (provide name or id, not both)
data "cumulocity_inventory_role" "by_id" {
  id = 1
}

output "restart_role_id" {
  value = data.cumulocity_inventory_role.restart.id
}

output "restart_permissions" {
  value = [for p in data.cumulocity_inventory_role.restart.permissions : p.permission]
}
