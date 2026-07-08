# All tenant options in a given category
data "cumulocity_tenant_options" "smtp" {
  category = "smtp"
}

# Values may contain credentials, so the whole option value is sensitive.
output "smtp_option_keys" {
  value = [for o in data.cumulocity_tenant_options.smtp.options : o.key]
}
