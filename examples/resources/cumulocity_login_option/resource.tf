resource "cumulocity_login_option" "oauth_internal" {
  type                   = "OAUTH2_INTERNAL"
  provider_name          = "Cumulocity"
  grant_type             = "PASSWORD"
  user_management_source = "INTERNAL"
  visible_on_login_page  = true
}

output "login_option_id" {
  value = cumulocity_login_option.oauth_internal.id
}
