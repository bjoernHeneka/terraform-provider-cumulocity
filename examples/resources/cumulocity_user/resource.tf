
# Beispiel 1: User mit Passwort
resource "cumulocity_user" "operator" {
  username   = "jdoe"
  email      = "johndoe@example.com"
  first_name = "John"
  last_name  = "Doe"
  password   = var.user_password
  enabled    = true
}

# Beispiel 2: User ohne Passwort — Cumulocity sendet eine Reset-Mail
resource "cumulocity_user" "new_admin" {
  username                  = "newadmin"
  email                     = "admin@example.com"
  send_password_reset_email = true
}

# Import eines bestehenden Users:
#   terraform import cumulocity_user.operator t0071234/jdoe
# oder (wenn tenant_id im Provider konfiguriert ist):
#   terraform import cumulocity_user.operator jdoe
