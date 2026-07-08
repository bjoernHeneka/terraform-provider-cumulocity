# Look up an application by name
data "cumulocity_application" "myapp" {
  name = "My Dashboard"
}

# Use the application ID in a subscription
resource "cumulocity_tenant_application_subscription" "example" {
  tenant_id      = "t0071234"
  application_id = data.cumulocity_application.myapp.id
}

# Output application details
output "application_id" {
  value = data.cumulocity_application.myapp.id
}

output "application_type" {
  value = data.cumulocity_application.myapp.type
}

output "owner_tenant" {
  value = data.cumulocity_application.myapp.owner_tenant_id
}
