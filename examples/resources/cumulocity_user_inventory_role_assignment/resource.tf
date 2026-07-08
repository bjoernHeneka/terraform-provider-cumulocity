resource "cumulocity_user" "alice" {
  username                  = "alice"
  email                     = "alice@example.com"
  send_password_reset_email = true
}

resource "cumulocity_managed_object" "device" {
  name      = "Field Device 1"
  type      = "c8y_Device"
  is_device = true
}

# Grant the user one or more inventory roles scoped to a specific managed object.
# Changing role_names updates the assignment in place.
resource "cumulocity_user_inventory_role_assignment" "alice_device" {
  user_id           = cumulocity_user.alice.username
  managed_object_id = cumulocity_managed_object.device.id
  role_names        = ["Operations: Restart Device"]
}
