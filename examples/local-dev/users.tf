
resource "cumulocity_user" "test" {
  username   = "tf-test-user"
  email      = "tf-test@codebee.de"
  first_name = "Terraform"
  last_name  = "Test"
  enabled    = true

  # Entweder Passwort direkt setzen:
  # password = var.new_user_password

  # Oder Cumulocity sendet eine Reset-Mail:
  send_password_reset_email = true
}
