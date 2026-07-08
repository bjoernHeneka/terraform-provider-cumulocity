---
page_title: "Data Source: cumulocity_login_option"
description: |-
  Looks up a single Cumulocity login option by its type or ID.
---

# cumulocity_login_option

Looks up a single Cumulocity login option (authentication configuration) by its type or ID. First-class scalar attributes cover the common fields; the complete, type-specific payload is available as raw JSON via `config_json`.

Corresponds to `GET /tenant/loginOptions/{typeOrId}`.

## Example Usage

```hcl
data "cumulocity_login_option" "oauth" {
  type_or_id = "OAUTH2"
}

output "login_option_issuer" {
  value = data.cumulocity_login_option.oauth.issuer
}
```

Deep, type-specific fields can be read from `config_json`. Because `config_json` is sensitive, `jsondecode()` yields a sensitive value; wrap only non-secret fields in `nonsensitive()`:

```hcl
locals {
  oauth_config = jsondecode(data.cumulocity_login_option.oauth.config_json)
}

output "token_request_url" {
  value = nonsensitive(local.oauth_config.tokenRequest.url)
}
```

## Schema

### Required

- `type_or_id` (String) — The login option type (e.g. `OAUTH2_INTERNAL`) or its ID to look up.

### Read-Only

- `id` (String) — The login option ID.
- `type` (String) — The authentication configuration type (e.g. `OAUTH2_INTERNAL`, `BASIC`).
- `provider_name` (String) — The name of the authentication provider.
- `grant_type` (String) — The OAuth2 grant type (e.g. `PASSWORD`, `AUTHORIZATION_CODE`).
- `user_management_source` (String) — The source of user management (e.g. `INTERNAL`).
- `visible_on_login_page` (Boolean) — Whether this login option is shown on the login page.
- `template` (String) — The configuration template, e.g. `CUSTOM` (OAuth2 options).
- `button_name` (String) — The label of the login button shown on the login page.
- `issuer` (String) — The OAuth2/OIDC token issuer URL.
- `client_id` (String) — The OAuth2 client ID.
- `audience` (String) — The OAuth2 token audience.
- `redirect_to_platform` (String) — The platform redirect URL used in the OAuth2 flow.
- `use_id_token` (Boolean) — Whether the ID token is used instead of the access token.
- `self` (String) — The self-link URL of the login option.
- `config_json` (String, Sensitive) — The complete raw JSON payload as returned by the API, including all type-specific nested fields (`tokenRequest`, `authorizationRequest`, `onNewUser`, `signatureVerificationConfig`, etc.). Parse with `jsondecode()`. Marked sensitive as request templates may contain secrets.
