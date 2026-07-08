---
page_title: "Resource: cumulocity_tenant_option"
description: |-
  Creates and manages a Cumulocity tenant option (category/key/value tuple).
---

# cumulocity_tenant_option

Creates and manages a Cumulocity tenant option. Tenant options are category/key/value tuples that store per-tenant configuration consumed by Cumulocity itself and by microservices.

- `category` and `key` are immutable — changing either forces a new resource.
- Only `value` can be updated in-place.
- Keys prefixed with `credentials.` cause the value to be stored encrypted. The `value` attribute is always marked sensitive.

Corresponds to `POST /tenant/options`, `GET/PUT/DELETE /tenant/options/{category}/{key}`.

## Example Usage

### CORS configuration

```hcl
resource "cumulocity_tenant_option" "cors" {
  category = "access.control"
  key      = "allow.origin"
  value    = "https://app.example.com,https://dashboard.example.com"
}
```

### Alarm type mapping

```hcl
resource "cumulocity_tenant_option" "alarm_temp" {
  category = "alarm.type.mapping"
  key      = "temp_too_high"
  value    = "CRITICAL|temperature too high"
}
```

### Encrypted credential (microservice config)

```hcl
resource "cumulocity_tenant_option" "api_secret" {
  category = "my-microservice"
  key      = "credentials.api-key"
  value    = var.api_key   # stored encrypted, never returned in plaintext
}
```

## Schema

### Required

- `category` (String) — Category of the option, e.g. `access.control`. Must not contain whitespace or special characters `$ & + , / : ; = ? @ " < > # % { } | \ ^ ~ [ ]`. Immutable — changing this value forces a new resource.
- `key` (String) — Key of the option, unique within its category. Prefix with `credentials.` to store the value encrypted. Immutable — changing this value forces a new resource.
- `value` (String, Sensitive) — Value of the option. This field is always marked sensitive. Options with a `credentials.` key prefix are stored encrypted in Cumulocity and the plaintext value is never returned by the API — Terraform preserves the last-applied value in state.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{category}/{key}`.
- `self` (String) — Self-link URL of the option.

## Import

Import an existing option using `{category}/{key}`.

```shell
terraform import cumulocity_tenant_option.cors access.control/allow.origin
```

~> **Note on credentials:** After importing an option with a `credentials.` key, the `value` attribute in state will be empty because the API returns `<<Encrypted>>`. Set the value explicitly in your configuration and run `terraform apply` to write it into state.
