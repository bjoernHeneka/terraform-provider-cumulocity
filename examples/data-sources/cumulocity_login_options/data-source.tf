# Retrieve all login options configured on the tenant
data "cumulocity_login_options" "all" {}

# All login option types on the tenant
output "login_option_types" {
  value = [for o in data.cumulocity_login_options.all.options : o.type]
}

# Only the options visible on the login page
output "visible_login_options" {
  value = [for o in data.cumulocity_login_options.all.options : o.type if o.visible_on_login_page]
}
