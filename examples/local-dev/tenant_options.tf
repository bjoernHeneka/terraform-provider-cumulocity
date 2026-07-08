
resource "cumulocity_tenant_option" "cors" {
  category = "access.control"
  key      = "allow.origin"
  value    = "*"
}

resource "cumulocity_tenant_option" "alarm_mapping" {
  category = "alarm.type.mapping"
  key      = "temp_too_high"
  value    = "CRITICAL|temperature too high"
}

# All options
data "cumulocity_tenant_options" "all" {}

# Only options in the "password" category
data "cumulocity_tenant_options" "password_opts" {
  category = "password"
}

output "all_option_keys" {
  value = [for o in data.cumulocity_tenant_options.all.options : "${o.category}/${o.key}"]
}

output "password_options" {
  value     = data.cumulocity_tenant_options.password_opts.options
  sensitive = true
}
