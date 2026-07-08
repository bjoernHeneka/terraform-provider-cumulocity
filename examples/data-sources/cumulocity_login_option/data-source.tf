# Look up a single login option by its type or ID
data "cumulocity_login_option" "cognito" {
  type_or_id = "OAUTH2"
}

# First-class scalar attributes
output "login_option_provider" {
  value = data.cumulocity_login_option.cognito.provider_name
}

output "login_option_issuer" {
  value = data.cumulocity_login_option.cognito.issuer
}

output "login_option_client_id" {
  value = data.cumulocity_login_option.cognito.client_id
}

# Deep, type-specific fields are available via config_json (raw API payload).
# config_json is sensitive, so jsondecode() yields a sensitive value and its
# sensitivity propagates. nonsensitive() is used below because these particular
# fields are not secret — never wrap fields that may contain tokens/secrets.
locals {
  cognito_config = jsondecode(data.cumulocity_login_option.cognito.config_json)
}

output "token_request_url" {
  value = nonsensitive(local.cognito_config.tokenRequest.url)
}

output "jwks_uri" {
  value = nonsensitive(local.cognito_config.signatureVerificationConfig.jwks.jwksUri)
}
