# A plain configuration option.
resource "cumulocity_tenant_option" "allow_origin" {
  category = "access.control"
  key      = "allow.origin"
  value    = "*"
}

# Keys prefixed with "credentials." are stored encrypted by Cumulocity.
resource "cumulocity_tenant_option" "smtp_password" {
  category = "smtp"
  key      = "credentials.password"
  value    = var.smtp_password
}

variable "smtp_password" {
  type      = string
  sensitive = true
}
