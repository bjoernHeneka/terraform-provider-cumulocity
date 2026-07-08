resource "cumulocity_tenant" "example" {
  company       = "ACME AG"
  domain        = "acme.cumulocity.com"
  admin_email   = "admin@acme.com"
  admin_name    = "acmeadmin"
  admin_pass    = "S3cur3P@ss!"
  contact_name  = "John Doe"
  contact_phone = "+49 123 456 7890"
}

output "tenant_id" {
  value = cumulocity_tenant.example.id
}

output "tenant_status" {
  value = cumulocity_tenant.example.status
}
