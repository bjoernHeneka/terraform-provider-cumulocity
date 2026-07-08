
resource "cumulocity_user_group" "operators" {
  name        = "tf-operators"
  description = "Managed by Terraform — operator access group"
}

resource "cumulocity_user_group_membership" "test_in_operators" {
  group_id = cumulocity_user_group.operators.group_id
  user_id  = cumulocity_user.test.username
}

output "operators_group_id" {
  value       = cumulocity_user_group.operators.group_id
  description = "Numeric group ID — use as group_id in cumulocity_user_group_membership"
}
