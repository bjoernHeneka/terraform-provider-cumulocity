# Look up a single global role by its exact name
data "cumulocity_role" "alarm_admin" {
  name = "ROLE_ALARM_ADMIN"
}

# Use the resolved role ID in a role assignment
resource "cumulocity_user_role_assignment" "example" {
  username = cumulocity_user.operator.username
  role_id  = data.cumulocity_role.alarm_admin.id
}

output "alarm_admin_self" {
  value = data.cumulocity_role.alarm_admin.self
}
