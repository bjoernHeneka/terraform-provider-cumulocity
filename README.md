# Terraform Provider for Cumulocity IoT

A Terraform provider for the [Cumulocity IoT platform](https://www.cumulocity.com/api).
Built on **terraform-plugin-framework** (not the legacy SDKv2).

| | |
|---|---|
| **Provider address** | `registry.terraform.io/bjoernHeneka/cumulocity` |
| **Go module** | `github.com/bjoernHeneka/terraform-provider-cumulocity` |
| **Go version** | 1.24+ |
| **terraform-plugin-framework** | v1.16+ |

---

## Table of Contents

- [Requirements](#requirements)
- [Provider Configuration](#provider-configuration)
- [Authentication](#authentication)
- [Resources](#resources)
  - [Identity & Access Control](#identity--access-control)
    - [cumulocity_user](#cumulocity_user)
    - [cumulocity_user_group](#cumulocity_user_group)
    - [cumulocity_user_group_membership](#cumulocity_user_group_membership)
    - [cumulocity_user_role_assignment](#cumulocity_user_role_assignment)
    - [cumulocity_user_inventory_role_assignment](#cumulocity_user_inventory_role_assignment)
  - [Inventory](#inventory)
    - [cumulocity_managed_object](#cumulocity_managed_object)
    - [cumulocity_external_id](#cumulocity_external_id)
    - [cumulocity_binary](#cumulocity_binary)
  - [Applications](#applications)
    - [cumulocity_application](#cumulocity_application)
    - [cumulocity_application_binary](#cumulocity_application_binary)
  - [Device Management](#device-management)
    - [cumulocity_device_operation](#cumulocity_device_operation)
    - [cumulocity_bulk_operation](#cumulocity_bulk_operation)
    - [cumulocity_new_device_request](#cumulocity_new_device_request)
    - [cumulocity_device_credentials](#cumulocity_device_credentials)
  - [Telemetry & Events](#telemetry--events)
    - [cumulocity_alarm](#cumulocity_alarm)
    - [cumulocity_event](#cumulocity_event)
    - [cumulocity_measurement](#cumulocity_measurement)
    - [cumulocity_audit_record](#cumulocity_audit_record)
  - [Tenant Administration](#tenant-administration)
    - [cumulocity_tenant](#cumulocity_tenant)
    - [cumulocity_tenant_application_subscription](#cumulocity_tenant_application_subscription)
    - [cumulocity_tenant_option](#cumulocity_tenant_option)
    - [cumulocity_trusted_certificate](#cumulocity_trusted_certificate)
    - [cumulocity_login_option](#cumulocity_login_option)
    - [cumulocity_login_option_raw](#cumulocity_login_option_raw)
  - [Platform Configuration](#platform-configuration)
    - [cumulocity_retention_rule](#cumulocity_retention_rule)
    - [cumulocity_notification_subscription](#cumulocity_notification_subscription)
- [Data Sources](#data-sources)
  - [cumulocity_application](#cumulocity_application-data-source)
  - [cumulocity_login_option](#cumulocity_login_option-data-source)
  - [cumulocity_login_options](#cumulocity_login_options-data-source)
  - [cumulocity_role](#cumulocity_role-data-source)
  - [cumulocity_roles](#cumulocity_roles-data-source)
  - [cumulocity_inventory_role](#cumulocity_inventory_role-data-source)
  - [cumulocity_inventory_roles](#cumulocity_inventory_roles-data-source)
  - [cumulocity_tenant_options](#cumulocity_tenant_options-data-source)
  - [cumulocity_operations](#cumulocity_operations-data-source)
  - [cumulocity_alarms](#cumulocity_alarms-data-source)
  - [cumulocity_events](#cumulocity_events-data-source)
  - [cumulocity_measurements](#cumulocity_measurements-data-source)
  - [cumulocity_managed_objects](#cumulocity_managed_objects-data-source)
  - [cumulocity_audit_records](#cumulocity_audit_records-data-source)
  - [cumulocity_binaries](#cumulocity_binaries-data-source)
- [Publishing the Provider](#publishing-the-provider)
- [Local Development](#local-development)
- [Running Tests](#running-tests)

---

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

---

## Provider Configuration

```hcl
terraform {
  required_providers {
    cumulocity = {
      source  = "registry.terraform.io/bjoernHeneka/cumulocity"
      version = "~> 0.1"
    }
  }
}

provider "cumulocity" {
  tenant_domain = "mytenant.cumulocity.com"
  tenant_id     = "t0071234"      # optional but recommended
  username      = "admin"
  password      = var.c8y_password
}
```

All four attributes can also be set via environment variables (environment variables take lower precedence than explicit configuration):

```bash
export CUMULOCITY_TENANT_DOMAIN=mytenant.cumulocity.com
export CUMULOCITY_TENANT_ID=t0071234
export CUMULOCITY_USERNAME=admin
export CUMULOCITY_PASSWORD=secret
```

### Provider Arguments

| Argument | Env var | Required | Description |
|---|---|---|---|
| `tenant_domain` | `CUMULOCITY_TENANT_DOMAIN` | Yes | Hostname of your Cumulocity instance, e.g. `mytenant.cumulocity.com` |
| `tenant_id` | `CUMULOCITY_TENANT_ID` | No | Short tenant ID, e.g. `t0071234`. Used to build Basic auth credentials |
| `username` | `CUMULOCITY_USERNAME` | Yes | Login username |
| `password` | `CUMULOCITY_PASSWORD` | Yes | Login password (sensitive) |

---

## Authentication

Cumulocity uses **HTTP Basic auth** with a tenant-scoped credential format:

```
Authorization: Basic base64(<tenantID>/<username>:<password>)
```

If `tenant_id` is omitted, the provider falls back to plain `username:password` Basic auth (suitable for single-tenant setups or when the tenant ID is embedded in the username).

The base URL is constructed as `https://<tenant_domain>`.

---

## Resources

### Identity & Access Control

---

#### `cumulocity_user`

Manages a Cumulocity user account.

**API:** `POST/GET/PUT/DELETE /user/{tenantId}/users/{userName}`

##### Example

```hcl
# User with a password
resource "cumulocity_user" "operator" {
  username   = "jdoe"
  email      = "johndoe@example.com"
  first_name = "John"
  last_name  = "Doe"
  password   = var.user_password
  enabled    = true
}

# User without a password — Cumulocity sends a password reset email
resource "cumulocity_user" "new_admin" {
  username                  = "newadmin"
  email                     = "admin@example.com"
  send_password_reset_email = true
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `username` | string | Yes | Login name. Cannot contain whitespace or `+$:/`. **Immutable** |
| `email` | string | Yes | Email address |
| `tenant_id` | string | No | Tenant to create the user in. Defaults to provider `tenant_id`. **Immutable** |
| `first_name` | string | No | First name |
| `last_name` | string | No | Last name |
| `display_name` | string | No | Display name shown in the UI |
| `phone` | string | No | Phone number |
| `enabled` | bool | No | Whether the account is enabled. Default: `true` |
| `newsletter` | bool | No | Subscribed to Cumulocity newsletter |
| `password` | string | No | Initial password (6–32 Latin-1 characters). **Sensitive. Write-only.** |
| `send_password_reset_email` | bool | No | Send a password reset email on creation instead of setting a password. Default: `false` |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{tenantId}/{userName}` |
| `self` | Self-link URI |
| `password_strength` | `GREEN`, `YELLOW`, or `RED` |
| `should_reset_password` | Whether the user must reset their password on next login |
| `last_password_change` | ISO 8601 timestamp of the last password change |

##### Import

```bash
terraform import cumulocity_user.operator t0071234/jdoe
# Or (if tenant_id is configured on the provider):
terraform import cumulocity_user.operator jdoe
```

---

#### `cumulocity_user_group`

Manages a user group within a tenant. Groups bundle roles and aggregate users for access control.

**API:** `POST/GET/PUT/DELETE /user/{tenantId}/groups/{groupId}`

##### Example

```hcl
resource "cumulocity_user_group" "device_managers" {
  name        = "DeviceManagers"
  description = "Users who can manage all devices"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Unique group name within the tenant |
| `tenant_id` | string | No | Defaults to provider `tenant_id`. **Immutable** |
| `description` | string | No | Free-text description |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{tenantId}/{groupId}` |
| `group_id` | Numeric group ID (used in `cumulocity_user_group_membership`) |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_user_group.device_managers t0071234/12345
# Or (if tenant_id is set on provider):
terraform import cumulocity_user_group.device_managers 12345
```

---

#### `cumulocity_user_group_membership`

Adds a user to a user group.

**API:** `POST/DELETE /user/{tenantId}/groups/{groupId}/users/{reference}`

##### Example

```hcl
resource "cumulocity_user_group_membership" "jdoe_device_managers" {
  username = cumulocity_user.operator.username
  group_id = cumulocity_user_group.device_managers.group_id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `username` | string | Yes | Username to add to the group. **Immutable** |
| `group_id` | number | Yes | Numeric group ID. **Immutable** |
| `tenant_id` | string | No | Defaults to provider `tenant_id`. **Immutable** |

##### Import

```bash
terraform import cumulocity_user_group_membership.example t0071234/12345/jdoe
```

---

#### `cumulocity_user_role_assignment`

Assigns a global role to a user.

**API:** `POST/DELETE /user/{tenantId}/users/{userId}/roles`

##### Example

```hcl
data "cumulocity_role" "device_control" {
  name = "ROLE_DEVICE_CONTROL_READ"
}

resource "cumulocity_user_role_assignment" "jdoe_device_control" {
  username = cumulocity_user.operator.username
  role_id  = data.cumulocity_role.device_control.id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `username` | string | Yes | Target user. **Immutable** |
| `role_id` | string | Yes | Role ID (from `cumulocity_role` data source). **Immutable** |
| `tenant_id` | string | No | Defaults to provider `tenant_id`. **Immutable** |

##### Import

```bash
terraform import cumulocity_user_role_assignment.example t0071234/jdoe/ROLE_DEVICE_CONTROL_READ
```

---

#### `cumulocity_user_inventory_role_assignment`

Assigns an inventory role to a user for a specific managed object.

**API:** `POST/DELETE /user/{tenantId}/users/{userId}/roles/inventory`

##### Example

```hcl
data "cumulocity_inventory_role" "reader" {
  name = "Reader"
}

resource "cumulocity_user_inventory_role_assignment" "jdoe_gateway" {
  username          = cumulocity_user.operator.username
  inventory_role_id = data.cumulocity_inventory_role.reader.id
  managed_object_id = cumulocity_managed_object.gateway.id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `username` | string | Yes | Target user. **Immutable** |
| `inventory_role_id` | number | Yes | Inventory role ID. **Immutable** |
| `managed_object_id` | string | Yes | Managed object ID. **Immutable** |
| `tenant_id` | string | No | Defaults to provider `tenant_id`. **Immutable** |

---

### Inventory

---

#### `cumulocity_managed_object`

Creates and manages a Cumulocity managed object — a device, device group, or generic asset in the inventory.

**API:** `POST/GET/PUT/DELETE /inventory/managedObjects/{id}`

##### Example

```hcl
# A device group
resource "cumulocity_managed_object" "factory_floor" {
  name            = "Factory Floor"
  type            = "c8y_DeviceGroup"
  is_device_group = true
}

# A device
resource "cumulocity_managed_object" "gateway" {
  name      = "Main Gateway"
  type      = "c8y_Gateway"
  is_device = true
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Display name |
| `type` | string | No | Device type class, e.g. `c8y_Gateway` |
| `is_device` | bool | No | Adds the `c8y_IsDevice` fragment |
| `is_device_group` | bool | No | Adds the `c8y_IsDeviceGroup` fragment |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Numeric managed object ID (assigned by Cumulocity) |
| `owner` | Username of the owner |
| `self` | Self-link URI |
| `creation_time` | ISO 8601 creation timestamp |
| `last_updated` | ISO 8601 last-update timestamp |

##### Import

```bash
terraform import cumulocity_managed_object.gateway 12345678
```

---

#### `cumulocity_external_id`

Links a managed object to an identifier in an external system (e.g. a device serial number).
All attributes are immutable — any change forces replacement.

**API:** `POST /identity/globalIds/{id}/externalIds`, `GET/DELETE /identity/externalIds/{type}/{externalId}`

##### Example

```hcl
resource "cumulocity_external_id" "serial" {
  type              = "c8y_Serial"
  external_id       = "SN-12345"
  managed_object_id = cumulocity_managed_object.gateway.id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | External ID type, e.g. `c8y_Serial`. **Immutable** |
| `external_id` | string | Yes | The identifier value in the external system. **Immutable** |
| `managed_object_id` | string | Yes | Managed object to link to. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{type}/{external_id}` |
| `managed_object_self` | Self-link of the linked managed object |
| `self` | Self-link of this external ID |

##### Import

```bash
terraform import cumulocity_external_id.serial c8y_Serial/SN-12345
```

---

#### `cumulocity_binary`

Uploads a file to the Cumulocity inventory binary store.

**API:** `POST /inventory/binaries`, `GET /inventory/managedObjects/{id}`, `DELETE /inventory/binaries/{id}`

##### Example

```hcl
resource "cumulocity_binary" "firmware" {
  file         = "${path.module}/firmware-v1.2.bin"
  file_hash    = filemd5("${path.module}/firmware-v1.2.bin")
  name         = "firmware-v1.2"
  content_type = "application/octet-stream"
}
```

> **Tip:** Use `file_hash = filemd5(...)` to trigger a re-upload when the file content changes but the path stays the same.

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `file` | string | Yes | Local filesystem path to the file. **Immutable** |
| `file_hash` | string | No | Hash of the file (e.g. `filemd5(...)`). When changed, forces re-upload. **Immutable** |
| `name` | string | No | Name for the binary object. Defaults to the base filename. **Immutable** |
| `content_type` | string | No | MIME type. Default: `application/octet-stream`. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Binary managed object ID |
| `length` | File size in bytes |
| `owner` | Owner username |
| `self` | Self-link URI |
| `last_updated` | ISO 8601 last-update timestamp |

##### Import

```bash
terraform import cumulocity_binary.firmware 12345678
```

---

### Applications

---

#### `cumulocity_application`

Creates and manages a Cumulocity application (hosted web app, external link, or microservice).

**API:** `POST/GET/PUT/DELETE /application/applications/{id}`

##### Example

```hcl
# Hosted web application
resource "cumulocity_application" "dashboard" {
  key          = "my-dashboard-key"
  name         = "My Dashboard"
  type         = "HOSTED"
  context_path = "my-dashboard"
  availability = "PRIVATE"
  description  = "Custom dashboard for factory floor"
}

# External link
resource "cumulocity_application" "docs" {
  key  = "external-docs-key"
  name = "API Docs"
  type = "EXTERNAL"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `key` | string | Yes | Unique application key across tenants |
| `name` | string | Yes | Display name |
| `type` | string | Yes | `HOSTED`, `EXTERNAL`, or `MICROSERVICE`. **Immutable** |
| `context_path` | string | No | URL path for HOSTED apps, e.g. `myapp` |
| `availability` | string | No | `MARKET` or `PRIVATE`. Default: `PRIVATE` |
| `description` | string | No | Human-readable description |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Application ID |
| `owner_tenant_id` | ID of the owning tenant |
| `active_version_id` | ID of the active binary version (set after `cumulocity_application_binary` upload) |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_application.dashboard 12345
```

---

#### `cumulocity_application_binary`

Uploads a ZIP archive as the binary for a `HOSTED` or `MICROSERVICE` application. All attributes are immutable.

**API:** `POST /application/applications/{id}/binaries`

##### Example

```hcl
resource "cumulocity_application_binary" "dashboard_v1" {
  application_id = cumulocity_application.dashboard.id
  file           = "${path.module}/dist/dashboard.zip"
  file_hash      = filemd5("${path.module}/dist/dashboard.zip")
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `application_id` | string | Yes | ID of the parent application. **Immutable** |
| `file` | string | Yes | Local path to the ZIP archive. **Immutable** |
| `file_hash` | string | No | File hash to detect content changes (e.g. `filemd5(...)`). **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Binary version ID |
| `version` | Version string extracted from the manifest |

---

### Device Management

---

#### `cumulocity_device_operation`

Sends an operation to a Cumulocity device.

The operation is created in `PENDING` status and progresses to `EXECUTING → SUCCESSFUL` or `FAILED` as the device processes it. On `terraform destroy`, `PENDING` operations are cancelled (set to `FAILED`); completed operations are simply removed from state.

All content attributes (`device_id`, `fragments_json`, `description`) are immutable — any change creates a new operation.

**API:** `POST/GET /devicecontrol/operations/{id}`

##### Example

```hcl
# Restart command
resource "cumulocity_device_operation" "restart" {
  device_id      = cumulocity_managed_object.gateway.id
  description    = "Restart gateway"
  fragments_json = jsonencode({ c8y_Restart = {} })
}

# Shell command
resource "cumulocity_device_operation" "list_files" {
  device_id      = cumulocity_managed_object.gateway.id
  fragments_json = jsonencode({
    c8y_Command = { text = "ls -la /etc" }
  })
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `device_id` | string | Yes | Target device managed object ID. **Immutable** |
| `fragments_json` | string | Yes | JSON object with the operation payload. **Immutable** |
| `description` | string | No | Human-readable description. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Operation ID |
| `status` | `PENDING`, `EXECUTING`, `SUCCESSFUL`, or `FAILED` |
| `failure_reason` | Populated when `status = FAILED` |
| `creation_time` | ISO 8601 creation timestamp |
| `bulk_operation_id` | ID of the parent bulk operation, if any |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_device_operation.restart 987654
```

---

#### `cumulocity_bulk_operation`

Creates a bulk operation that sends the same operation to multiple devices matching a filter.

**API:** `POST/GET /devicecontrol/bulkoperations/{id}`

##### Example

```hcl
resource "cumulocity_bulk_operation" "firmware_update" {
  description         = "Firmware update v2.0 for all gateways"
  start_time          = "2024-06-01T08:00:00.000Z"
  creation_ramp       = 15
  operation_prototype = jsonencode({
    c8y_Firmware = {
      name    = "gateway-firmware"
      version = "2.0.0"
      url     = "https://firmware.example.com/v2.0.0.bin"
    }
  })
  group_id = cumulocity_managed_object.factory_floor.id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `description` | string | No | Human-readable description |
| `start_time` | string | Yes | ISO 8601 scheduled start time |
| `creation_ramp` | number | Yes | Rate at which operations are created (ms between each) |
| `operation_prototype` | string | Yes | JSON object for the operation payload |
| `group_id` | string | No | Target device group managed object ID |
| `failed_parent_id` | string | No | Re-executes operations that failed in a previous bulk operation |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Bulk operation ID |
| `status` | `SCHEDULED`, `EXECUTING`, `EXECUTED`, or `CANCELLED` |
| `progress` | Object with `all`, `pending`, `executing`, `successful`, `failed` counts |

---

#### `cumulocity_new_device_request`

Registers a new device request. Used in the device bootstrap process before device credentials are issued.

**API:** `POST/GET/DELETE /devicecontrol/newDeviceRequests/{requestId}`

##### Example

```hcl
resource "cumulocity_new_device_request" "sensor_001" {
  request_id = "sensor-001-serial"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `request_id` | string | Yes | Unique identifier for the new device request. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Same as `request_id` |
| `status` | `WAITING_FOR_CONNECTION` or `PENDING_ACCEPTANCE` |
| `self` | Self-link URI |

---

#### `cumulocity_device_credentials`

Accepts pending device credentials for a registered device request.

**API:** `POST /devicecontrol/deviceCredentials`

##### Example

```hcl
resource "cumulocity_device_credentials" "sensor_001" {
  request_id = cumulocity_new_device_request.sensor_001.request_id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `request_id` | string | Yes | The device request ID to accept credentials for. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{tenantId}/{username}` |
| `tenant_id` | Tenant that issued the credentials |
| `username` | Generated username for the device |
| `password` | Generated password for the device (**Sensitive**) |
| `self` | Self-link URI |

---

### Telemetry & Events

---

#### `cumulocity_alarm`

Creates and manages an alarm on a managed object.

**API:** `POST/GET/PUT/DELETE /alarm/alarms/{id}`

##### Example

```hcl
resource "cumulocity_alarm" "high_temperature" {
  source_id = cumulocity_managed_object.gateway.id
  type      = "c8y_TemperatureAlarm"
  text      = "Temperature exceeded threshold"
  severity  = "MAJOR"
  time      = "2024-01-15T10:30:00.000Z"
}

# Acknowledge an alarm
resource "cumulocity_alarm" "connection_lost" {
  source_id = cumulocity_managed_object.gateway.id
  type      = "c8y_UnavailabilityAlarm"
  text      = "Device unreachable"
  severity  = "CRITICAL"
  status    = "ACKNOWLEDGED"
  time      = "2024-01-15T09:00:00.000Z"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `source_id` | string | Yes | Managed object ID the alarm is associated with. **Immutable** |
| `type` | string | Yes | Alarm type, e.g. `c8y_UnavailabilityAlarm`. **Immutable** |
| `text` | string | Yes | Human-readable description |
| `severity` | string | Yes | `CRITICAL`, `MAJOR`, `MINOR`, or `WARNING`. **Immutable** |
| `time` | string | Yes | ISO 8601 occurrence time. **Immutable** |
| `status` | string | No | `ACTIVE`, `ACKNOWLEDGED`, or `CLEARED`. Defaults to `ACTIVE` |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Alarm ID |
| `occurrence_count` | Times this alarm has been deduplicated |
| `creation_time` | ISO 8601 creation timestamp |
| `last_updated` | ISO 8601 last-update timestamp |
| `first_occurrence_time` | ISO 8601 first-occurrence timestamp (when count > 1) |

##### Import

```bash
terraform import cumulocity_alarm.high_temperature 12345678
```

---

#### `cumulocity_event`

Creates and manages a time-stamped event on a managed object.

Only `text` can be updated in-place; all other attributes are immutable.

**API:** `POST/GET/PUT/DELETE /event/events/{id}`

##### Example

```hcl
resource "cumulocity_event" "location_update" {
  source_id = cumulocity_managed_object.gateway.id
  type      = "c8y_LocationUpdate"
  text      = "Device reported new GPS coordinates"
  time      = "2024-01-15T10:30:00.000Z"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `source_id` | string | Yes | Managed object ID. **Immutable** |
| `type` | string | Yes | Event type, e.g. `c8y_LocationUpdate`. **Immutable** |
| `text` | string | Yes | Human-readable description |
| `time` | string | Yes | ISO 8601 occurrence time. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Event ID |
| `creation_time` | ISO 8601 creation timestamp |
| `last_updated` | ISO 8601 last-update timestamp |

##### Import

```bash
terraform import cumulocity_event.location_update 12345678
```

---

#### `cumulocity_measurement`

Creates a measurement. Measurements are immutable — any attribute change forces replacement.

**API:** `POST/GET/DELETE /measurement/measurements/{id}`

##### Example

```hcl
resource "cumulocity_measurement" "temperature" {
  source_id = cumulocity_managed_object.gateway.id
  type      = "c8y_TemperatureMeasurement"
  time      = "2024-01-15T10:30:00.000Z"
  fragments = jsonencode({
    c8y_Temperature = {
      T = { value = 22.5, unit = "°C" }
    }
  })
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `source_id` | string | Yes | Managed object ID. **Immutable** |
| `type` | string | Yes | Measurement type. **Immutable** |
| `time` | string | Yes | ISO 8601 measurement time. **Immutable** |
| `fragments` | string | No | JSON object with measurement fragment data. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Measurement ID |
| `creation_time` | ISO 8601 creation timestamp |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_measurement.temperature 12345678
```

---

#### `cumulocity_audit_record`

Creates an audit record for compliance or custom audit trail purposes.

**API:** `POST/GET /audit/auditRecords/{id}`

##### Example

```hcl
resource "cumulocity_audit_record" "config_change" {
  type       = "Configuration"
  text       = "Updated MQTT broker address"
  source_id  = cumulocity_managed_object.gateway.id
  activity   = "Update"
  severity   = "INFORMATION"
  time       = "2024-01-15T11:00:00.000Z"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Audit record type |
| `text` | string | Yes | Description of the audited action |
| `source_id` | string | Yes | Associated managed object ID. **Immutable** |
| `activity` | string | Yes | The audited activity, e.g. `Update`, `Delete` |
| `severity` | string | Yes | `INFORMATION`, `WARNING`, or `CRITICAL` |
| `time` | string | Yes | ISO 8601 time of the audit event |

---

### Tenant Administration

---

#### `cumulocity_tenant`

Creates and manages a Cumulocity subtenant.

**Required roles:** `ROLE_TENANT_MANAGEMENT_ADMIN` or `ROLE_TENANT_MANAGEMENT_CREATE`

**API:** `POST/GET/PUT/DELETE /tenant/tenants/{tenantId}`

##### Example

```hcl
resource "cumulocity_tenant" "subsidiary" {
  company      = "Acme Corp"
  domain       = "acme.cumulocity.com"
  admin_email  = "admin@acme.com"
  admin_name   = "acmeadmin"
  admin_pass   = var.acme_admin_password
  contact_name = "Jane Doe"
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `company` | string | Yes | Company name |
| `domain` | string | Yes | Tenant domain. **Immutable** |
| `admin_email` | string | Yes | Admin email address |
| `admin_name` | string | No | Admin username. **Immutable** |
| `admin_pass` | string | No | Admin password. **Sensitive. Write-only** |
| `contact_name` | string | No | Contact person name |
| `contact_phone` | string | No | Contact phone in international format |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Tenant ID, e.g. `t0071234` |
| `parent` | Parent tenant ID |
| `status` | `ACTIVE` or `SUSPENDED` |
| `creation_time` | RFC 3339 creation timestamp |
| `allow_create_tenants` | Whether this tenant can create subtenants |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_tenant.subsidiary t0099999
```

---

#### `cumulocity_tenant_application_subscription`

Subscribes a tenant to an application. This allows a subtenant to use an application owned by the parent tenant or another tenant.

**Required roles:** `ROLE_APPLICATION_MANAGEMENT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`/`UPDATE`

**API:** `POST/DELETE /tenant/tenants/{tenantId}/applications`

##### Example

```hcl
# Subscribe a subtenant to an application
resource "cumulocity_tenant_application_subscription" "myapp_subscription" {
  tenant_id      = cumulocity_tenant.subsidiary.id
  application_id = cumulocity_application.myapp.id
}

# Subscribe using direct IDs
resource "cumulocity_tenant_application_subscription" "example" {
  tenant_id      = "t0071234"
  application_id = "12345"
}

# Multiple subscriptions for the same tenant
resource "cumulocity_tenant" "subtenant" {
  company     = "Example Corp"
  domain      = "example.cumulocity.com"
  admin_email = "admin@example.com"
}

resource "cumulocity_application" "dashboard" {
  name         = "Dashboard App"
  type         = "HOSTED"
  key          = "dashboard-key"
  availability = "PRIVATE"
}

resource "cumulocity_application" "reports" {
  name         = "Reports App"
  type         = "HOSTED"
  key          = "reports-key"
  availability = "PRIVATE"
}

resource "cumulocity_tenant_application_subscription" "dashboard_sub" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = cumulocity_application.dashboard.id
}

resource "cumulocity_tenant_application_subscription" "reports_sub" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = cumulocity_application.reports.id
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `tenant_id` | string | Yes | The tenant ID to subscribe (e.g. `t0071234`). **Immutable** |
| `application_id` | string | Yes | The application ID to subscribe to. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{tenantId}/{applicationId}` |

##### Import

```bash
terraform import cumulocity_tenant_application_subscription.example t0071234/12345
```

> **Note:** Subscription requests are asynchronous. A successful create indicates the request was accepted, but Kubernetes resources may still be provisioning. Similarly, unsubscribe returns immediately, but resource cleanup continues in the background.

---

#### `cumulocity_tenant_option`

Manages a tenant option (a category/key/value configuration tuple).

Tenant options store per-tenant configuration consumed by Cumulocity and microservices. Keys prefixed with `credentials.` are stored encrypted.

**API:** `POST/GET/PUT/DELETE /tenant/options/{category}/{key}`

##### Example

```hcl
resource "cumulocity_tenant_option" "cors" {
  category = "access.control"
  key      = "allow.origin"
  value    = "https://myapp.example.com"
}

# Encrypted credential (credentials.* keys are stored encrypted)
resource "cumulocity_tenant_option" "smtp_password" {
  category = "email"
  key      = "credentials.smtp.password"
  value    = var.smtp_password
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `category` | string | Yes | Option category. **Immutable** |
| `key` | string | Yes | Option key. Prefix with `credentials.` for encrypted storage. **Immutable** |
| `value` | string | Yes | Option value. **Sensitive** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{category}/{key}` |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_tenant_option.cors "access.control/allow.origin"
```

---

#### `cumulocity_trusted_certificate`

Uploads a trusted X.509 certificate for a tenant. Devices use these certificates to establish connections.

The PEM content is immutable — changing it forces replacement.

**API:** `POST/GET/PUT/DELETE /tenant/tenants/{tenantId}/trusted-certificates/{fingerprint}`

##### Example

```hcl
resource "cumulocity_trusted_certificate" "device_ca" {
  cert_in_pem_format       = file("${path.module}/ca.pem")
  status                   = "ENABLED"
  name                     = "Device CA"
  auto_registration_enabled = true
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `cert_in_pem_format` | string | Yes | X.509 certificate in PEM format. **Immutable** |
| `status` | string | Yes | `ENABLED` or `DISABLED` |
| `name` | string | No | Human-readable name |
| `auto_registration_enabled` | bool | No | Allow devices to auto-register with this certificate |
| `tenant_id` | string | No | Tenant to upload to. Defaults to provider `tenant_id`. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Composite: `{tenantId}/{fingerprint}` |
| `fingerprint` | Certificate fingerprint |
| `algorithm_name` | Signing algorithm |
| `issuer` | Organization that signed the certificate |
| `not_before` / `not_after` | Certificate validity window |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_trusted_certificate.device_ca "t0071234/AB:CD:EF:..."
```

---

#### `cumulocity_login_option`

Creates and manages a tenant authentication/login option.

**Required roles:** `ROLE_TENANT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`

**API:** `POST/GET/PUT/DELETE /tenant/loginOptions/{typeOrId}`

##### Example

```hcl
resource "cumulocity_login_option" "basic" {
  type                = "BASIC"
  provider_name       = "Cumulocity"
  visible_on_login_page = true
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | Auth type: `BASIC`, `OAUTH2`, or `OAUTH2_INTERNAL`. **Immutable** |
| `provider_name` | string | Yes | Display name shown in the UI |
| `grant_type` | string | No | OAuth grant type: `AUTHORIZATION_CODE` or `PASSWORD` |
| `user_management_source` | string | No | `INTERNAL` or `REMOTE` |
| `visible_on_login_page` | bool | No | Show on the login page |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Login option ID |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_login_option.basic BASIC
```

---

#### `cumulocity_login_option_raw`

Manages a login option via a verbatim JSON body. Use this instead of `cumulocity_login_option` when you need to manage the full `authConfig`, including nested/type-specific fields such as `tokenRequest`, `authorizationRequest`, `onNewUser.dynamicMapping`, or `signatureVerificationConfig` (typical for OAUTH2 / SSO providers).

**Required roles:** `ROLE_TENANT_ADMIN` or `ROLE_TENANT_MANAGEMENT_ADMIN`

**API:** `POST/GET/PUT/DELETE /tenant/loginOptions`

##### Example

```hcl
resource "cumulocity_login_option_raw" "oauth" {
  body = jsonencode({
    type         = "OAUTH2"
    providerName = "Corporate SSO"
    grantType    = "AUTHORIZATION_CODE"
    tokenRequest = {
      url = "https://sso.example.com/oauth2/token"
    }
    authorizationRequest = {
      url = "https://sso.example.com/oauth2/authorize"
    }
  })
}
```

Build `body` with `jsonencode({...})` so formatting stays stable. This resource does not reconcile individual fields against the server — drift detection is limited to whether the option still exists. The server masks secrets (for example, `client_secret`) in responses, so `config_json` is not a round-trippable copy of `body`.

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `body` | string | Yes | The complete `authConfig` as a JSON string, sent verbatim to the API |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Login option ID assigned by Cumulocity |
| `config_json` | Raw JSON payload as returned by the API after the last create/update/read (**Sensitive**) |

##### Import

```bash
terraform import cumulocity_login_option_raw.oauth OAUTH2
```

After importing, define `body` in your configuration to match the desired state; the `config_json` output can serve as a starting point once masked secret placeholders (e.g. `"****"`) are replaced with real values. `body` itself is not populated by import.

---

### Platform Configuration

---

#### `cumulocity_retention_rule`

Manages a data retention rule controlling how long specific types of data are kept.

**API:** `POST/GET/PUT/DELETE /retention/retentions/{id}`

##### Example

```hcl
# Keep alarms for 30 days
resource "cumulocity_retention_rule" "alarms" {
  data_type   = "ALARM"
  maximum_age = 30
}

# Keep temperature measurements for 90 days
resource "cumulocity_retention_rule" "temperature" {
  data_type     = "MEASUREMENT"
  fragment_type = "c8y_TemperatureMeasurement"
  maximum_age   = 90
}

# Catch-all: keep everything else for 365 days
resource "cumulocity_retention_rule" "default" {
  data_type   = "*"
  maximum_age = 365
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `maximum_age` | number | Yes | Retention period in days |
| `data_type` | string | No | `ALARM`, `AUDIT`, `BULK_OPERATION`, `EVENT`, `MEASUREMENT`, `OPERATION`, or `*`. Default: `*` |
| `fragment_type` | string | No | Fragment type filter. Applies to `EVENT`, `MEASUREMENT`, `OPERATION`, `BULK_OPERATION`. Default: `*` |
| `source` | string | No | Source filter. Default: `*` |
| `type` | string | No | Type filter. Applies to `ALARM`, `AUDIT`, `EVENT`, `MEASUREMENT`. Default: `*` |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Retention rule ID |
| `editable` | `false` for system-managed rules |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_retention_rule.alarms 12345
```

---

#### `cumulocity_notification_subscription`

Manages a Notification 2.0 subscription that defines which data is forwarded for a device or tenant context.

All attributes are immutable — any change forces replacement.

**API:** `POST/GET/DELETE /notification2/subscriptions/{id}`

##### Example

```hcl
# Subscribe to alarms and events for a specific device
resource "cumulocity_notification_subscription" "device_alerts" {
  context      = "mo"
  source_id    = cumulocity_managed_object.gateway.id
  subscription = "GatewayAlerts"
  apis         = ["alarms", "events"]
  type_filter  = "'c8y_UnavailabilityAlarm' or 'c8y_ConnectivityAlarm'"
}

# Tenant-level inventory subscription
resource "cumulocity_notification_subscription" "tenant_inventory" {
  context      = "tenant"
  subscription = "TenantInventory"
  apis         = ["managedobjects"]
}

# All data for a device, forwarding only specific fragments
resource "cumulocity_notification_subscription" "device_full" {
  context           = "mo"
  source_id         = cumulocity_managed_object.sensor.id
  subscription      = "SensorFull"
  apis              = ["*"]
  fragments_to_copy = ["c8y_Temperature", "c8y_Position"]
  non_persistent    = false
}
```

##### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `context` | string | Yes | `mo` (managed object) or `tenant`. **Immutable** |
| `subscription` | string | Yes | Subscription name (alphanumeric only). **Immutable** |
| `source_id` | string | No | Managed object ID. Required when `context = "mo"`. **Immutable** |
| `apis` | list(string) | No | APIs to subscribe to: `alarms`, `alarmsWithChildren`, `events`, `eventsWithChildren`, `managedobjects`, `measurements`, `operations`, `*`. **Immutable** |
| `type_filter` | string | No | OData type filter, e.g. `'c8y_Speed' or 'c8y_LocationUpdate'`. **Immutable** |
| `fragments_to_copy` | list(string) | No | Custom fragment names to include. **Immutable** |
| `non_persistent` | bool | No | When `true`, messages may be lost if no consumer is connected. Default: `false`. **Immutable** |

##### Attributes Reference (Computed)

| Attribute | Description |
|---|---|
| `id` | Subscription ID |
| `self` | Self-link URI |

##### Import

```bash
terraform import cumulocity_notification_subscription.device_alerts 12345678
```

---

## Data Sources

### `cumulocity_application` (Data Source)

Looks up a single application by name.

```hcl
data "cumulocity_application" "myapp" {
  name = "My Dashboard"
}

# Use in a tenant application subscription:
resource "cumulocity_tenant_application_subscription" "example" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = data.cumulocity_application.myapp.id
}

# Output application details
output "app_info" {
  value = {
    id    = data.cumulocity_application.myapp.id
    type  = data.cumulocity_application.myapp.type
    owner = data.cumulocity_application.myapp.owner_tenant_id
  }
}
```

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Application name to look up |

| Attribute | Description |
|---|---|
| `id` | Application ID |
| `key` | Application key |
| `type` | Application type: `HOSTED`, `EXTERNAL`, or `MICROSERVICE` |
| `context_path` | Context path (for HOSTED applications) |
| `availability` | `MARKET` or `PRIVATE` |
| `description` | Application description |
| `active_version_id` | ID of the active binary version |
| `owner_tenant_id` | ID of the owning tenant |
| `self` | Self-link URI |

> **Note:** If multiple applications with the same name exist, the data source will fail. Application names should be unique for data source lookups.

---

### `cumulocity_login_option` (Data Source)

Looks up a single login option by its type or ID. First-class scalar attributes cover the common fields; the complete, type-specific payload is available as raw JSON via `config_json`.

```hcl
data "cumulocity_login_option" "oauth" {
  type_or_id = "OAUTH2"
}

output "login_option_issuer" {
  value = data.cumulocity_login_option.oauth.issuer
}
```

| Argument | Required | Description |
|---|---|---|
| `type_or_id` | Yes | Login option type (e.g. `OAUTH2_INTERNAL`) or ID to look up |

| Attribute | Description |
|---|---|
| `id` | Login option ID |
| `type` | Auth configuration type (e.g. `OAUTH2_INTERNAL`, `BASIC`) |
| `provider_name` | Name of the authentication provider |
| `grant_type` | OAuth2 grant type (e.g. `PASSWORD`, `AUTHORIZATION_CODE`) |
| `user_management_source` | Source of user management (e.g. `INTERNAL`) |
| `visible_on_login_page` | Whether the option is shown on the login page |
| `template` | Configuration template, e.g. `CUSTOM` |
| `button_name` | Label of the login button |
| `issuer` | OAuth2/OIDC token issuer URL |
| `client_id` | OAuth2 client ID |
| `audience` | OAuth2 token audience |
| `redirect_to_platform` | Platform redirect URL used in the OAuth2 flow |
| `use_id_token` | Whether the ID token is used instead of the access token |
| `self` | Self-link URI |
| `config_json` | Complete raw JSON payload, including type-specific nested fields; parse with `jsondecode()` (**Sensitive**) |

---

### `cumulocity_login_options` (Data Source)

Returns all login options configured on the tenant. Each entry exposes the same fields as `cumulocity_login_option`.

```hcl
data "cumulocity_login_options" "all" {}

output "login_option_types" {
  value = [for o in data.cumulocity_login_options.all.options : o.type]
}
```

| Attribute | Description |
|---|---|
| `options` | List of login option objects (same fields as `cumulocity_login_option`) |

---

### `cumulocity_role` (Data Source)

Looks up a single global role by name.

```hcl
data "cumulocity_role" "device_control" {
  name = "ROLE_DEVICE_CONTROL_READ"
}

# Use in a role assignment:
resource "cumulocity_user_role_assignment" "example" {
  username = cumulocity_user.operator.username
  role_id  = data.cumulocity_role.device_control.id
}
```

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Role name to look up |

| Attribute | Description |
|---|---|
| `id` | Role ID |
| `self` | Self-link URI |

---

### `cumulocity_roles` (Data Source)

Returns all global roles available in the tenant.

```hcl
data "cumulocity_roles" "all" {}

output "role_names" {
  value = [for r in data.cumulocity_roles.all.roles : r.name]
}
```

| Attribute | Description |
|---|---|
| `roles` | List of `{ id, name, self }` objects |

---

### `cumulocity_inventory_role` (Data Source)

Looks up a single inventory role by name.

```hcl
data "cumulocity_inventory_role" "reader" {
  name = "Reader"
}
```

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Inventory role name to look up |

| Attribute | Description |
|---|---|
| `id` | Numeric inventory role ID |
| `description` | Role description |
| `self` | Self-link URI |

---

### `cumulocity_inventory_roles` (Data Source)

Returns all inventory roles.

```hcl
data "cumulocity_inventory_roles" "all" {}
```

| Attribute | Description |
|---|---|
| `roles` | List of `{ id, name, description, self }` objects |

---

### `cumulocity_tenant_options` (Data Source)

Returns tenant options, optionally filtered by category.

```hcl
data "cumulocity_tenant_options" "access" {
  category = "access.control"
}

output "allowed_origins" {
  value = { for o in data.cumulocity_tenant_options.access.options : o.key => o.value }
}
```

| Argument | Required | Description |
|---|---|---|
| `category` | No | Filter by category |

| Attribute | Description |
|---|---|
| `options` | List of `{ category, key, value, self }` objects |

---

### `cumulocity_operations` (Data Source)

Returns device operations, optionally filtered by device ID and/or status.

```hcl
data "cumulocity_operations" "pending_gateway" {
  device_id = cumulocity_managed_object.gateway.id
  status    = "PENDING"
}
```

| Argument | Required | Description |
|---|---|---|
| `device_id` | No | Filter by target device |
| `status` | No | `PENDING`, `EXECUTING`, `SUCCESSFUL`, or `FAILED` |

| Attribute | Description |
|---|---|
| `operations` | List of operation objects |

---

### `cumulocity_alarms` (Data Source)

Returns alarms matching optional filters.

```hcl
data "cumulocity_alarms" "critical" {
  source_id = cumulocity_managed_object.gateway.id
  severity  = "CRITICAL"
  status    = "ACTIVE"
}
```

| Argument | Required | Description |
|---|---|---|
| `source_id` | No | Filter by managed object |
| `type` | No | Filter by alarm type |
| `severity` | No | `CRITICAL`, `MAJOR`, `MINOR`, or `WARNING` |
| `status` | No | `ACTIVE`, `ACKNOWLEDGED`, or `CLEARED` |

| Attribute | Description |
|---|---|
| `alarms` | List of alarm objects |

---

### `cumulocity_events` (Data Source)

Returns events matching optional filters.

```hcl
data "cumulocity_events" "location_updates" {
  source_id = cumulocity_managed_object.gateway.id
  type      = "c8y_LocationUpdate"
}
```

| Argument | Required | Description |
|---|---|---|
| `source_id` | No | Filter by managed object |
| `type` | No | Filter by event type |
| `date_from` | No | ISO 8601 lower bound |
| `date_to` | No | ISO 8601 upper bound |

| Attribute | Description |
|---|---|
| `events` | List of event objects |

---

### `cumulocity_measurements` (Data Source)

Returns measurements matching optional filters.

```hcl
data "cumulocity_measurements" "recent_temp" {
  source_id     = cumulocity_managed_object.gateway.id
  type          = "c8y_TemperatureMeasurement"
  date_from     = "2024-01-01T00:00:00.000Z"
}
```

| Argument | Required | Description |
|---|---|---|
| `source_id` | No | Filter by managed object |
| `type` | No | Filter by measurement type |
| `date_from` / `date_to` | No | ISO 8601 time bounds |

| Attribute | Description |
|---|---|
| `measurements` | List of measurement objects |

---

### `cumulocity_managed_objects` (Data Source)

Returns managed objects matching optional filters.

```hcl
data "cumulocity_managed_objects" "gateways" {
  type = "c8y_Gateway"
}
```

| Argument | Required | Description |
|---|---|---|
| `type` | No | Filter by object type |
| `name` | No | Filter by name (supports wildcards) |
| `fragment_type` | No | Filter by fragment presence |
| `owner` | No | Filter by owner username |
| `query` | No | Raw inventory query string |

| Attribute | Description |
|---|---|
| `managed_objects` | List of managed object objects |

---

### `cumulocity_audit_records` (Data Source)

Returns audit records matching optional filters.

```hcl
data "cumulocity_audit_records" "config_changes" {
  type     = "Configuration"
  source   = cumulocity_managed_object.gateway.id
}
```

| Argument | Required | Description |
|---|---|---|
| `type` | No | Filter by audit record type |
| `source` | No | Filter by associated managed object ID |
| `user` | No | Filter by username who performed the action |
| `date_from` / `date_to` | No | ISO 8601 time bounds |

| Attribute | Description |
|---|---|
| `audit_records` | List of audit record objects |

---

### `cumulocity_binaries` (Data Source)

Returns binaries from the inventory binary store, optionally filtered by owner or type. All pages are followed automatically.

```hcl
data "cumulocity_binaries" "mine" {
  owner = "admin"
}
```

| Argument | Required | Description |
|---|---|---|
| `owner` | No | Filter by the username of the binary owner |
| `type` | No | Filter by the managed object type of the binary |

| Attribute | Description |
|---|---|
| `binaries` | List of binary managed object objects |

---

## Publishing the Provider

The public Terraform Registry and the HCP Terraform private registry use **different publishing mechanisms**. Pick the one you need — the release tooling in this repository (`.goreleaser.yml`, `terraform-registry-manifest.json`, `.github/workflows/release.yml`) is required for both.

Every release must be a Git tag that is a valid semantic version **with a leading `v`** (e.g. `v0.1.0`). Tags without the `v` prefix are ignored by the registry. Never modify or re-tag a published version — publish a new one instead.

### One-time setup: GPG signing key

Release checksums must be GPG-signed:

```bash
# Generate a key (RSA 4096, no expiry)
gpg --full-generate-key

# Export the public key (registered with the registry)
gpg --armor --export <KEY_ID> > public_key.asc

# Export the private key (stored as a GitHub Actions secret)
gpg --armor --export-secret-keys <KEY_ID>
```

Add two GitHub Actions secrets to the repository: `GPG_PRIVATE_KEY` (armored private key) and `GPG_PASSPHRASE`.

### Cutting a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs GoReleaser, which builds the per-platform zips and attaches the four asset types the registry requires to the GitHub release:

- `terraform-provider-cumulocity_{VERSION}_{os}_{arch}.zip` (one per platform)
- `terraform-provider-cumulocity_{VERSION}_SHA256SUMS`
- `terraform-provider-cumulocity_{VERSION}_SHA256SUMS.sig` (binary, detached GPG signature)
- `terraform-provider-cumulocity_{VERSION}_manifest.json`

**A release without these assets is invisible to every registry.**

### Option A — Public Terraform Registry (registry.terraform.io)

The public registry ingests GitHub releases automatically once the provider is connected:

1. The GitHub repository must be public and named `terraform-provider-{name}` (this one is).
2. Sign in at [registry.terraform.io](https://registry.terraform.io) with the GitHub account/org that owns the repo.
3. **Publish → Provider**, select the repository, and add the GPG public key (`public_key.asc`) when prompted.
4. The registry picks up the latest valid release and ingests new tags automatically via webhook.

The provider source address is then `bjoernHeneka/cumulocity`.

### Option B — HCP Terraform private registry (app.terraform.io)

**The private registry does not watch GitHub.** Pushing tags or creating GitHub releases has no effect on it — every version must be pushed through the [Registry Providers API](https://developer.hashicorp.com/terraform/cloud-docs/registry/publish-providers).

One-time: register the GPG public key with the private registry and note the returned `key-id`:

```bash
curl -sS \
  --header "Authorization: Bearer $TFC_TOKEN" \
  --header "Content-Type: application/vnd.api+json" \
  --request POST \
  --data "{\"data\":{\"type\":\"gpg-keys\",\"attributes\":{\"namespace\":\"<YOUR_ORG>\",\"ascii-armor\":$(jq -Rs . < public_key.asc)}}}" \
  https://app.terraform.io/api/registry/private/v2/gpg-keys
```

Per release: build the assets, then run the publishing script (creates the provider record, version, platforms, and uploads all files):

```bash
goreleaser release --clean        # or download the assets of an existing GitHub release into ./dist

TFC_TOKEN=<api-token> \
TFC_ORG=<YOUR_ORG> \
GPG_KEY_ID=<key-id> \
VERSION=0.1.0 \
./scripts/publish-tfc-private-registry.sh
```

Use it in configurations with the private source address:

```hcl
terraform {
  required_providers {
    cumulocity = {
      source  = "app.terraform.io/<YOUR_ORG>/cumulocity"
      version = "~> 0.1"
    }
  }
}
```

### Troubleshooting: provider does not show up

| Symptom | Cause |
|---|---|
| Public registry does not list a version | The GitHub release has no GoReleaser assets (tag was pushed before the release workflow/secrets existed), or the tag lacks the `v` prefix. Fix the tooling, then publish a **new** tag. |
| HCP Terraform private registry shows nothing | The private registry never ingests from GitHub — run the API publishing flow (Option B). |
| Version exists but `terraform init` fails checksum verification | A published release was modified or re-tagged. Publish a new version. |

---

## Local Development

### Use from GitHub (without publishing to a registry)

The easiest way to use the provider locally is to build it directly from the GitHub repository and point Terraform to the binary via a dev override.

**1. Clone and build**

```bash
git clone https://github.com/bjoernHeneka/terraform-provider-cumulocity
cd terraform-provider-cumulocity
make install
```

`make install` builds the provider and places the binary in `$(go env GOPATH)/bin` (typically `~/go/bin`).

**2. Configure `~/.terraformrc`**

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/bjoernHeneka/cumulocity" = "/Users/<you>/go/bin"
  }
  direct {}
}
```

Replace `/Users/<you>/go/bin` with the output of `go env GOPATH`/bin on your machine.

**3. Use the provider in your Terraform config**

```hcl
terraform {
  required_providers {
    cumulocity = {
      source = "registry.terraform.io/bjoernHeneka/cumulocity"
    }
  }
}
```

**4. Skip `terraform init` — use directly**

With dev overrides active, Terraform bypasses the registry lookup. You can run `terraform plan` and `terraform apply` directly without running `terraform init` first.

Terraform will show a warning — this is expected:

```
│ Warning: Provider development overrides are in effect
│ The following provider development overrides are set in the CLI configuration:
│  - registry.terraform.io/bjoernHeneka/cumulocity in /Users/<you>/go/bin
```

**After every code change:**

```bash
make install   # rebuild
terraform plan # picks up the new binary immediately
```

---

### Build only (without installing)

```bash
go build ./...
```

### Install for local testing (dev override)

```bash
make install
```

This installs the binary to `~/go/bin/terraform-provider-cumulocity`.

Configure `~/.terraformrc` to use it:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/bjoernHeneka/cumulocity" = "/Users/<you>/go/bin"
  }
  direct {}
}
```

With this in place, `terraform plan/apply` uses your locally built binary without needing `terraform init`.

---

## Running Tests

### Unit tests

```bash
go test ./internal/...
```

### Acceptance tests (requires a real Cumulocity tenant)

```bash
export CUMULOCITY_TENANT_DOMAIN=mytenant.cumulocity.com
export CUMULOCITY_TENANT_ID=t0071234
export CUMULOCITY_USERNAME=admin
export CUMULOCITY_PASSWORD=secret

TF_ACC=1 go test -v -timeout 120m ./internal/provider/...
```

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add your resource in `internal/provider/<name>_resource.go` and the corresponding client methods in `internal/client/`
4. Register the resource in `provider.go`
5. Add an example in `examples/resources/cumulocity_<name>/resource.tf`
6. Run `go mod tidy` and `go build ./...`
7. Open a pull request

See the [`docs/`](docs/) directory for the full provider reference, including per-resource and per-data-source pages.
