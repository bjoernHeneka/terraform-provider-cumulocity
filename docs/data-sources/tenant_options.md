---
page_title: "Data Source: cumulocity_tenant_options"
description: |-
  Retrieves all tenant options, optionally filtered by category.
---

# cumulocity_tenant_options

Retrieves all tenant options for the current tenant, with an optional `category` filter. All pages are followed automatically.

Use this data source to inspect existing configuration, feed option values into other resources, or audit the full set of tenant options.

Corresponds to `GET /tenant/options` (paginated).

## Example Usage

### All options

```hcl
data "cumulocity_tenant_options" "all" {}

output "all_option_keys" {
  value = [for o in data.cumulocity_tenant_options.all.options : "${o.category}/${o.key}"]
}
```

### Filter by category

```hcl
data "cumulocity_tenant_options" "access_control" {
  category = "access.control"
}

output "allowed_origins" {
  value     = [for o in data.cumulocity_tenant_options.access_control.options : o.value if o.key == "allow.origin"]
  sensitive = true
}
```

### Pass an option value to another resource

```hcl
data "cumulocity_tenant_options" "password" {
  category = "password"
}

locals {
  password_options = { for o in data.cumulocity_tenant_options.password.options : o.key => o.value }
}

output "password_validity" {
  value     = local.password_options["limit.validity"]
  sensitive = true
}
```

## Schema

### Optional

- `category` (String) — When set, only options belonging to this category are returned. When omitted, all options from all categories are returned.

### Read-Only

- `options` (List of Object) — The matching tenant options. Each object has:
  - `category` (String) — Category of the option.
  - `key` (String) — Key of the option.
  - `value` (String, Sensitive) — Value of the option.
  - `self` (String) — Self-link URL of the option.
