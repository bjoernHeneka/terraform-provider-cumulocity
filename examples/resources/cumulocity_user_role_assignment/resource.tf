resource "cumulocity_user" "alice" {
  username                  = "alice"
  email                     = "alice@example.com"
  send_password_reset_email = true
}

# Assign a global role to the user.
resource "cumulocity_user_role_assignment" "alice_alarm_admin" {
  user_id = cumulocity_user.alice.username
  role    = "ROLE_ALARM_ADMIN"
}

resource "cumulocity_user_role_assignment" "alice_inventory_read" {
  user_id = cumulocity_user.alice.username
  role    = "ROLE_INVENTORY_READ"
}
