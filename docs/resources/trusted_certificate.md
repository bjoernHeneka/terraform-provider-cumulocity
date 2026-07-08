---
page_title: "Resource: cumulocity_trusted_certificate"
description: |-
  Uploads and manages a trusted X.509 certificate for a Cumulocity tenant.
---

# cumulocity_trusted_certificate

Uploads and manages a trusted X.509 certificate for a Cumulocity tenant. Devices use these certificates to establish connections with the platform.

Corresponds to `POST/GET/PUT/DELETE /tenant/tenants/{tenantId}/trusted-certificates/{fingerprint}`.

## Example Usage

```hcl
resource "cumulocity_trusted_certificate" "example" {
  # tenant_id defaults to the provider's tenant_id if omitted
  name                      = "My Device CA"
  status                    = "ENABLED"
  auto_registration_enabled = true
  cert_in_pem_format        = file("${path.module}/ca.pem")
}

output "fingerprint" {
  value = cumulocity_trusted_certificate.example.fingerprint
}
```

## Schema

### Required

- `cert_in_pem_format` (String) — The trusted certificate in PEM format. Changing this forces a new resource.
- `status` (String) — Whether the certificate is active: `ENABLED` or `DISABLED`.

### Optional

- `tenant_id` (String) — The tenant ID to upload the certificate to. Defaults to the provider's `tenant_id`. Changing this forces a new resource.
- `name` (String) — Human-readable name for the certificate.
- `auto_registration_enabled` (Boolean) — Whether devices can auto-register using this certificate.

### Read-Only

- `id` (String) — Composite Terraform identifier: `{tenantId}/{fingerprint}`.
- `fingerprint` (String) — Unique fingerprint of the certificate (assigned by Cumulocity).
- `algorithm_name` (String) — Algorithm used to encode the certificate.
- `issuer` (String) — The organization that signed the certificate.
- `not_after` (String) — End of the certificate's validity period.
- `not_before` (String) — Start of the certificate's validity period.
- `self` (String) — Self-link URL of the trusted certificate.

## Import

Import an existing trusted certificate using `{tenantId}/{fingerprint}`:

```shell
terraform import cumulocity_trusted_certificate.example t0071234/a1b2c3d4e5f6
```

~> **Note:** The `cert_in_pem_format` attribute is not populated on import, as the API may not return the PEM body on read. Set it in your configuration after importing.
