---
page_title: "Resource: cumulocity_login_option_raw"
description: |-
  Manages a Cumulocity login option via a verbatim JSON body.
---

# cumulocity_login_option_raw

Manages a Cumulocity login option (authentication configuration) via a verbatim JSON body. Use this instead of [`cumulocity_login_option`](login_option.md) when you need to manage the full `authConfig`, including nested/type-specific fields such as `tokenRequest`, `authorizationRequest`, `onNewUser.dynamicMapping` or `signatureVerificationConfig` (e.g. OAUTH2 / SSO providers).

Requires `ROLE_TENANT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`.

Corresponds to `POST/GET/PUT/DELETE /tenant/loginOptions`.

~> **Note on drift:** This resource does not reconcile individual fields against the server — drift detection is limited to whether the option still exists. The server masks secrets (e.g. `client_secret`) in responses, so `config_json` is not a round-trippable copy of `body`.

## Example Usage

```hcl
# The real OAuth2 client secret. The API masks it as "****" in responses, so it
# must be supplied here. Pass via TF_VAR_cognito_client_secret or a *.tfvars file.
variable "cognito_client_secret" {
  type      = string
  sensitive = true
}

# Full OAUTH2 (Cognito) login option managed as a verbatim JSON body.
#
# IMPORTANT: inside jsonencode() strings, `${...}` is Terraform interpolation.
# Cumulocity request-template placeholders must therefore be escaped as `$${...}`
# so Terraform emits a literal `${...}` to the API. Real interpolation (the
# client secret) uses a single `${...}`.
resource "cumulocity_login_option_raw" "cognito" {
  body = jsonencode({
    type                 = "OAUTH2"
    template             = "CUSTOM"
    providerName         = "cognito"
    buttonName           = "Login with Cognito"
    userManagementSource = "REMOTE"
    visibleOnLoginPage   = true
    grantType            = "AUTHORIZATION_CODE"
    useIdToken           = true

    issuer   = "https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_example"
    audience = "ClientID"
    clientId = "ClientID"

    tokenRequest = {
      method    = "POST"
      operation = "EXECUTE"
      url       = "https://example.auth.eu-central-1.amazoncognito.com/oauth2/token"
      body      = "grant_type=authorization_code&code=$${code}&client_id=$${clientId}&client_secret=${var.cognito_client_secret}"
    }
  })
}

output "cognito_id" {
  value = cumulocity_login_option_raw.cognito.id
}

# The server-side payload (with normalized/added fields and masked secrets).
output "cognito_effective_config" {
  value     = jsondecode(cumulocity_login_option_raw.cognito.config_json)
  sensitive = true
}
```

## Schema

### Required

- `body` (String) — The complete `authConfig` as a JSON string, sent verbatim to the API. Use `jsonencode({...})` to build it from an HCL object so formatting stays stable.

### Read-Only

- `id` (String) — Unique identifier of the login option assigned by Cumulocity.
- `config_json` (String, Sensitive) — The complete raw JSON payload as returned by the API after the last create/update/read, including server-added and normalized fields. Parse with `jsondecode()`. Marked sensitive as it may echo request templates.

## Import

Import an existing login option by its ID (or type):

```shell
terraform import cumulocity_login_option_raw.cognito oauth2-cognito
```

~> **Note:** `body` is not populated by import. After importing, define `body` in your configuration to match the desired state. You can use the `config_json` output as a starting point, replacing any masked secret placeholders (e.g. `"****"`) with real values.
