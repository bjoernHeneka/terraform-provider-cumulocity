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
