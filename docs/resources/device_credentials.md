---
page_title: "Resource: cumulocity_device_credentials"
description: |-
  Requests auto-generated credentials for a Cumulocity device.
---

# cumulocity_device_credentials

Requests auto-generated username and password for a device identified by its external ID. The credentials are returned only on creation — they are stored in Terraform state and can be passed to device provisioning workflows.

**Lifecycle note:** Destroy removes the resource from Terraform state but the credentials persist in Cumulocity. They are tied to the device's lifecycle and cannot be individually deleted via the API.

> **⚠️ Bootstrap credentials required**
>
> `POST /devicecontrol/deviceCredentials` is restricted by Cumulocity to the **bootstrap user** (`devicebootstrap`). Calling it with a regular admin account returns HTTP 403.
>
> To use this resource, configure the provider with bootstrap credentials:
> ```hcl
> provider "cumulocity" {
>   tenant_domain = "mytenant.cumulocity.com"
>   username      = "devicebootstrap"
>   password      = var.bootstrap_password
> }
> ```
> The bootstrap user credentials can be found in the Cumulocity UI under **Device Management → Device credentials → Device bootstrap credentials**.

Corresponds to `POST /devicecontrol/deviceCredentials`.

## Example Usage

```hcl
resource "cumulocity_new_device_request" "gateway" {
  device_id = "factory-gateway-01"
  status    = "ACCEPTED"
}

resource "cumulocity_device_credentials" "gateway" {
  device_id = cumulocity_new_device_request.gateway.device_id
}

output "gateway_username" {
  value = cumulocity_device_credentials.gateway.username
}

output "gateway_password" {
  value     = cumulocity_device_credentials.gateway.password
  sensitive = true
}
```

### Pass credentials to a remote device via SSH

```hcl
resource "null_resource" "provision_device" {
  triggers = {
    device_id = cumulocity_device_credentials.gateway.device_id
  }

  provisioner "remote-exec" {
    connection {
      host = var.device_ip
    }
    inline = [
      "tedge config set c8y.url ${var.tenant_domain}",
      "tedge connect c8y --username '${cumulocity_device_credentials.gateway.username}'",
    ]
  }
}
```

## Schema

### Required

- `device_id` (String) — External ID of the device. Changing this value forces a new resource.

### Optional

- `security_token` (String, Sensitive) — One-time security token. Changing this value forces a new resource.

### Read-Only

- `id` (String) — Terraform identifier (equals `device_id`).
- `username` (String) — Auto-generated device username.
- `password` (String, Sensitive) — Auto-generated device password. Only available immediately after creation.
- `tenant_id` (String) — Tenant ID associated with these credentials.
- `self` (String) — Self-link URL.

## Import

Import by device external ID. Note that `username` and `password` will not be populated after import — the API does not expose them after initial creation.

```shell
terraform import cumulocity_device_credentials.gateway factory-gateway-01
```
