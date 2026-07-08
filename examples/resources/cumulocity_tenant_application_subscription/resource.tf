# Subscribe a tenant to an application using direct IDs
resource "cumulocity_tenant_application_subscription" "example" {
  tenant_id      = "t0071234"
  application_id = "12345"
}

# Subscribe using references from other resources
resource "cumulocity_tenant" "subtenant" {
  company     = "Example Corp"
  domain      = "example.cumulocity.com"
  admin_email = "admin@example.com"
}

resource "cumulocity_application" "myapp" {
  name         = "my-application"
  type         = "HOSTED"
  key          = "my-app-key"
  availability = "PRIVATE"
}

resource "cumulocity_tenant_application_subscription" "sub" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = cumulocity_application.myapp.id
}

# Subscribe to an existing application using data source
data "cumulocity_application" "existing_app" {
  name = "Existing Dashboard"
}

resource "cumulocity_tenant_application_subscription" "existing" {
  tenant_id      = cumulocity_tenant.subtenant.id
  application_id = data.cumulocity_application.existing_app.id
}
