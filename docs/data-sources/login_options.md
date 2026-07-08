---
page_title: "Data Source: cumulocity_login_options"
description: |-
  Lists all Cumulocity login options configured on the tenant.
---

# cumulocity_login_options

Retrieves all login options (authentication configurations) configured on the tenant. Each entry exposes the same fields as the `cumulocity_login_option` data source, including the raw `config_json` payload.

Corresponds to `GET /tenant/loginOptions`.

## Example Usage

```hcl
data "cumulocity_login_options" "all" {}

# All login option types on the tenant
output "login_option_types" {
  value = [for o in data.cumulocity_login_options.all.options : o.type]
}

# Only the options visible on the login page
output "visible_login_options" {
  value = [for o in data.cumulocity_login_options.all.options : o.type if o.visible_on_login_page]
}
```

## Schema

### Read-Only

- `options` (List of Object) — List of all login options on the tenant. Each object contains:
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
  - `config_json` (String, Sensitive) — The complete raw JSON payload as returned by the API, including all type-specific nested fields. Parse with `jsondecode()`. Marked sensitive as request templates may contain secrets.
