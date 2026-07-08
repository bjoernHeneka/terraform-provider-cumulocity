
resource "cumulocity_user_inventory_role_assignment" "test" {
  user_id           = cumulocity_user.test.username
  managed_object_id = cumulocity_managed_object.test_device.id

  role_names = [
    data.cumulocity_inventory_role.restart.name,
  ]
}

output "test_inventory_assignment_id" {
  value = cumulocity_user_inventory_role_assignment.test.assignment_id
}
