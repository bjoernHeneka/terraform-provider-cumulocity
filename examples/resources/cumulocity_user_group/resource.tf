# A user group within the provider's default tenant.
resource "cumulocity_user_group" "operators" {
  name        = "operators"
  description = "Field operators with device control permissions."
}

# Use the computed group_id when adding members via
# cumulocity_user_group_membership.
output "operators_group_id" {
  value = cumulocity_user_group.operators.group_id
}
