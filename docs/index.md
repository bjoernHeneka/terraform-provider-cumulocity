---
page_title: "Provider: Cumulocity"
description: |-
  The Cumulocity provider manages users, roles, and inventory objects in a Cumulocity IoT tenant via the Cumulocity REST API.
---

# Cumulocity Provider

The Cumulocity provider lets you manage resources in a [Cumulocity IoT](https://www.cumulocity.com/) tenant declaratively with Terraform.

## Authentication

Cumulocity uses **Basic authentication** with a tenant-scoped credential:

```
Authorization: Basic base64(<tenantId>/<username>:<password>)
```

All four values can be supplied in the provider block or via environment variables.

The provider validates the supplied credentials at configuration time by calling the
current-user endpoint. Misconfiguration therefore fails fast during `terraform plan`
rather than on the first resource operation. Unknown configuration values (for example,
an attribute derived from another resource's output) are rejected with a hint to use the
corresponding environment variable instead. `tenant_domain` must be a bare hostname —
without a URL scheme or trailing path — because the base URL is constructed as
`https://<tenant_domain>`.

### Credential handling

- Credentials are only ever sent to `https://<tenant_domain>`.
- Connections use a minimum of TLS 1.2.
- `password` is marked sensitive and is never written to logs or plan output.
- Prefer environment variables or a secrets manager over committing credentials to
  plaintext `.tfvars` or the provider block.

## Example Usage

```hcl
terraform {
  required_providers {
    cumulocity = {
      source  = "registry.terraform.io/org-codebee/cumulocity"
      version = "~> 0.1"
    }
  }
}

provider "cumulocity" {
  tenant_domain = "mytenant.cumulocity.com"
  tenant_id     = "t0071234"
  username      = "admin"
  password      = var.cumulocity_password
}
```

## Schema

All attributes are declared optional in the provider schema, but `tenant_domain`,
`username`, and `password` must be supplied either in the provider block or via the
corresponding environment variable — configuration fails if any of them is missing.
`tenant_domain` must be a bare hostname, without a URL scheme or trailing path.

### Required (via block or environment variable)

- `tenant_domain` (String) — The Cumulocity tenant domain, e.g. `mytenant.cumulocity.com`. Can also be set via the `CUMULOCITY_TENANT_DOMAIN` environment variable.
- `username` (String) — Username for Basic auth. Can also be set via `CUMULOCITY_USERNAME`.
- `password` (String, Sensitive) — Password for Basic auth. Can also be set via `CUMULOCITY_PASSWORD`.

### Optional

- `tenant_id` (String) — The short tenant ID, e.g. `t0071234`. Used to build the Basic auth credential (`<tenantId>/<username>:<password>`). Can also be set via `CUMULOCITY_TENANT_ID`.

## Environment Variables

| Variable                    | Provider attribute |
|-----------------------------|--------------------|
| `CUMULOCITY_TENANT_DOMAIN`  | `tenant_domain`    |
| `CUMULOCITY_TENANT_ID`      | `tenant_id`        |
| `CUMULOCITY_USERNAME`       | `username`         |
| `CUMULOCITY_PASSWORD`       | `password`         |

## Resources

| Resource | Description |
|----------|-------------|
| [cumulocity_alarm](resources/alarm.md) | Manages a Cumulocity alarm on a device or asset |
| [cumulocity_audit_record](resources/audit_record.md) | Creates an immutable audit record for a platform action |
| [cumulocity_event](resources/event.md) | Manages a time-stamped event on a device or asset |
| [cumulocity_user](resources/user.md) | Manages a Cumulocity user account |
| [cumulocity_user_role_assignment](resources/user_role_assignment.md) | Assigns a global role to a user |
| [cumulocity_user_inventory_role_assignment](resources/user_inventory_role_assignment.md) | Assigns inventory roles to a user for a specific managed object |
| [cumulocity_managed_object](resources/managed_object.md) | Manages a device, group, or asset in the Cumulocity inventory |
| [cumulocity_external_id](resources/external_id.md) | Maps an external identifier to a managed object |
| [cumulocity_binary](resources/binary.md) | Uploads a file to the inventory binary store |
| [cumulocity_user_group](resources/user_group.md) | Manages a user group within a tenant |
| [cumulocity_user_group_membership](resources/user_group_membership.md) | Assigns a user to a user group |
| [cumulocity_application](resources/application.md) | Manages a Cumulocity application (HOSTED, EXTERNAL, MICROSERVICE) |
| [cumulocity_application_binary](resources/application_binary.md) | Uploads a ZIP archive to an application |
| [cumulocity_device_operation](resources/device_operation.md) | Sends an operation command to a device |
| [cumulocity_new_device_request](resources/new_device_request.md) | Manages a device registration request |
| [cumulocity_device_credentials](resources/device_credentials.md) | Requests auto-generated credentials for a device |
| [cumulocity_bulk_operation](resources/bulk_operation.md) | Sends a bulk operation to all devices in a group |
| [cumulocity_measurement](resources/measurement.md) | Manages a measurement on a device or asset |
| [cumulocity_tenant](resources/tenant.md) | Manages a Cumulocity subtenant |
| [cumulocity_tenant_option](resources/tenant_option.md) | Manages a tenant configuration option (category/key/value) |
| [cumulocity_tenant_application_subscription](resources/tenant_application_subscription.md) | Subscribes a tenant to an application |
| [cumulocity_trusted_certificate](resources/trusted_certificate.md) | Manages a trusted device certificate for a tenant |
| [cumulocity_login_option](resources/login_option.md) | Manages a tenant authentication/login option |
| [cumulocity_login_option_raw](resources/login_option_raw.md) | Manages a login option via a verbatim JSON body (full authConfig) |
| [cumulocity_retention_rule](resources/retention_rule.md) | Manages a data retention rule |
| [cumulocity_notification_subscription](resources/notification_subscription.md) | Manages a Notification 2.0 subscription |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [cumulocity_role](data-sources/role.md) | Looks up a single global role by name |
| [cumulocity_roles](data-sources/roles.md) | Lists all global roles, with optional name filter |
| [cumulocity_inventory_role](data-sources/inventory_role.md) | Looks up a single inventory role by name or ID |
| [cumulocity_inventory_roles](data-sources/inventory_roles.md) | Lists all inventory roles, with optional name filter |
| [cumulocity_tenant_options](data-sources/tenant_options.md) | Lists all tenant options, with optional category filter |
| [cumulocity_operations](data-sources/operations.md) | Lists device operations with optional device ID and status filters |
| [cumulocity_alarms](data-sources/alarms.md) | Lists alarms with optional source, status, severity, and type filters |
| [cumulocity_audit_records](data-sources/audit_records.md) | Lists audit records with optional source, type, user, and application filters |
| [cumulocity_events](data-sources/events.md) | Lists events with optional source, type, and date range filters |
| [cumulocity_measurements](data-sources/measurements.md) | Lists measurements with optional source, type, date range, and fragment series filters |
| [cumulocity_managed_objects](data-sources/managed_objects.md) | Lists managed objects with optional type, fragment, query, text, and owner filters |
| [cumulocity_binaries](data-sources/binaries.md) | Lists inventory binaries, with optional owner and type filters |
| [cumulocity_application](data-sources/application.md) | Looks up a single application by name |
| [cumulocity_login_option](data-sources/login_option.md) | Looks up a single login option by type or ID |
| [cumulocity_login_options](data-sources/login_options.md) | Lists all login options configured on the tenant |
