resource "cumulocity_user_group" "operators" {
  name        = "operators"
  description = "Field operators."
}

resource "cumulocity_user" "alice" {
  username                  = "alice"
  email                     = "alice@example.com"
  send_password_reset_email = true
}

# Add the user to the group. group_id references the computed numeric group ID.
resource "cumulocity_user_group_membership" "alice_operator" {
  group_id = cumulocity_user_group.operators.group_id
  user_id  = cumulocity_user.alice.username
}
