---
page_title: "Resource: cumulocity_new_device_request"
description: |-
  Creates and manages a Cumulocity device registration request.
---

# cumulocity_new_device_request

Creates and manages a Cumulocity device registration request. This is the first step in the standard device onboarding flow:

1. Create a `cumulocity_new_device_request` with `device_id` = the device's external ID.
2. The device bootstraps and receives `WAITING_FOR_CONNECTION` status.
3. Once the device establishes a connection, the status changes to `PENDING_ACCEPTANCE`.
4. Set `status = "ACCEPTED"` in your config and run `terraform apply` to approve the device.
5. Optionally create [`cumulocity_device_credentials`](device_credentials.md) to get the auto-generated login credentials.

`device_id` is immutable — changing it forces a new resource. `status` and `security_token` can be updated in-place (e.g. to move the request to `ACCEPTED`). `group_id` and `device_type` are applied on create only; change them by recreating the resource.

Corresponds to `POST/GET/PUT/DELETE /devicecontrol/newDeviceRequests/{requestId}`.

## Example Usage

### Register and auto-accept a device

```hcl
resource "cumulocity_new_device_request" "gateway" {
  device_id = "my-gateway-001"
  status    = "ACCEPTED"
}
```

### Full registration flow with group assignment

```hcl
resource "cumulocity_new_device_request" "sensor" {
  device_id   = "temp-sensor-42"
  group_id    = cumulocity_managed_object.factory_floor.id
  device_type = "c8y_TemperatureSensor"
  status      = "ACCEPTED"
}

resource "cumulocity_device_credentials" "sensor" {
  device_id = cumulocity_new_device_request.sensor.device_id
}

output "sensor_credentials" {
  value = {
    username = cumulocity_device_credentials.sensor.username
    password = cumulocity_device_credentials.sensor.password
  }
  sensitive = true
}
```

### Two-phase acceptance with security token

```hcl
resource "cumulocity_new_device_request" "secure_device" {
  device_id      = "secure-device-007"
  status         = "ACCEPTED"
  security_token = var.device_security_token
}
```

## Schema

### Required

- `device_id` (String) — External ID of the device (used as the request identifier). Immutable — changing this value forces a new resource.

### Optional

- `status` (String) — Registration status to set: `WAITING_FOR_CONNECTION`, `PENDING_ACCEPTANCE`, or `ACCEPTED`. Set to `ACCEPTED` to approve the device.
- `group_id` (String) — ID of the device group to assign the device to upon acceptance.
- `device_type` (String) — Device type, e.g. `c8y_Linux`.
- `security_token` (String, Sensitive) — Security token to verify against the device's token. Required when accepting a device if the security token policy is enforced. Write-only.

### Read-Only

- `id` (String) — Terraform identifier (equals `device_id`).
- `tenant_id` (String) — Tenant that owns this request.
- `owner` (String) — Username of the request owner.
- `creation_time` (String) — ISO 8601 timestamp when the request was created.
- `self` (String) — Self-link URL.

## Import

Import by the device external ID.

```shell
terraform import cumulocity_new_device_request.gateway my-gateway-001
```
