---
page_title: "Resource: cumulocity_user"
description: |-
  Creates and manages a Cumulocity user account.
---

# cumulocity_user

Creates and manages a Cumulocity user account.

Corresponds to `POST/GET/PUT/DELETE /user/{tenantId}/users/{userName}`.

## Example Usage

### User with password

```hcl
resource "cumulocity_user" "alice" {
  username   = "alice"
  email      = "alice@example.com"
  first_name = "Alice"
  last_name  = "Example"
  enabled    = true
  password   = var.alice_password
}
```

### User with password reset email on first login

```hcl
resource "cumulocity_user" "bob" {
  username                  = "bob"
  email                     = "bob@example.com"
  send_password_reset_email = true
}
```

## Schema

### Required

- `username` (String) — Login name of the user. Immutable — changing this value forces a new resource.
- `email` (String) — Email address of the user.

### Optional

- `tenant_id` (String) — Cumulocity tenant ID. Defaults to the provider's `tenant_id`. Changing this value forces a new resource.
- `first_name` (String) — First name.
- `last_name` (String) — Last name.
- `display_name` (String) — Display name shown in the UI.
- `phone` (String) — Phone number.
- `enabled` (Boolean) — Whether the user account is active. Defaults to `true`.
- `newsletter` (Boolean) — Whether the user receives newsletters.
- `password` (String, Sensitive) — User's password. Must be 6–32 characters. Required on create if `send_password_reset_email` is not set. This field is write-only — it is never read back from the API and is stored in state only.
- `send_password_reset_email` (Boolean) — When `true`, sends a password reset email to the user. Required on create if `password` is not set.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{userName}`.
- `self` (String) — Self-link URL of the user resource.
- `password_strength` (String) — Strength of the current password: `GREEN`, `YELLOW`, or `RED`.
- `should_reset_password` (Boolean) — Whether the user is required to reset their password on next login.
- `last_password_change` (String) — ISO 8601 timestamp of the most recent password change.

## Import

Import an existing user using `{tenantId}/{userName}`. If `tenantId` is omitted, the provider's configured `tenant_id` is used.

```shell
terraform import cumulocity_user.alice t0071234/alice
# or, using the provider's default tenant
terraform import cumulocity_user.alice alice
```

~> **Note on password:** Because `password` is write-only, it will not be populated after import. Terraform will not detect drift on this field. Set `send_password_reset_email = true` or explicitly set the password attribute after importing.
