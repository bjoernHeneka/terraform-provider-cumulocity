
resource "cumulocity_user_role_assignment" "test_alarm_admin" {
  user_id = cumulocity_user.test.username
  role    = "ROLE_ALARM_ADMIN"
}

resource "cumulocity_user_role_assignment" "test_device_control" {
  user_id = cumulocity_user.test.username
  role    = "ROLE_DEVICE_CONTROL_ADMIN"
}
