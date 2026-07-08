# List all inventory roles
data "cumulocity_inventory_roles" "all" {}

# Narrow the results to roles whose name contains "Operations"
data "cumulocity_inventory_roles" "operations" {
  name_filter = "Operations"
}

output "all_inventory_role_names" {
  value = [for r in data.cumulocity_inventory_roles.all.roles : r.name]
}

output "operations_role_ids" {
  value = [for r in data.cumulocity_inventory_roles.operations.roles : r.id]
}
