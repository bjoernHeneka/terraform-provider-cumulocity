---
page_title: "Resource: cumulocity_login_option"
description: |-
  Manages a Cumulocity login option (authentication configuration).
---

# cumulocity_login_option

Creates and manages a Cumulocity login option (authentication configuration). This resource manages the core fields of a login option. To manage the full `authConfig`, including nested/type-specific fields (e.g. OAUTH2/SSO providers), use [`cumulocity_login_option_raw`](login_option_raw.md).

Requires `ROLE_TENANT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`.

Corresponds to `POST/GET/PUT/DELETE /tenant/loginOptions/{typeOrId}`.

## Example Usage

```hcl
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
```

## Schema

### Required

- `type` (String) — Authentication type, e.g. `BASIC`, `OAUTH2`, `OAUTH2_INTERNAL`. Changing this forces a new resource.
- `provider_name` (String) — Display name of the authentication provider shown in the UI.

### Optional

- `grant_type` (String) — OAuth grant type: `AUTHORIZATION_CODE` or `PASSWORD`.
- `user_management_source` (String) — Source of user management, e.g. `INTERNAL` or `REMOTE`.
- `visible_on_login_page` (Boolean) — Whether this login option is shown on the login page.

### Read-Only

- `id` (String) — Unique identifier of the login option assigned by Cumulocity.
- `self` (String) — Self-link URL of the login option.
- `template` (String) — The configuration template, e.g. `CUSTOM` (OAuth2 options).
- `button_name` (String) — The label of the login button shown on the login page.
- `issuer` (String) — The OAuth2/OIDC token issuer URL.
- `client_id` (String) — The OAuth2 client ID.
- `audience` (String) — The OAuth2 token audience.
- `redirect_to_platform` (String) — The platform redirect URL used in the OAuth2 flow.
- `use_id_token` (Boolean) — Whether the ID token is used instead of the access token.
- `config_json` (String, Sensitive) — The complete raw JSON payload as returned by the API, including all type-specific nested fields. Parse with `jsondecode()`. Marked sensitive as request templates may contain secrets.

## Import

Import an existing login option by its ID:

```shell
terraform import cumulocity_login_option.oauth_internal oauth2-internal
```
