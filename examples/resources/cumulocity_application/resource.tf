# A hosted web application served by Cumulocity.
# Upload its ZIP archive with cumulocity_application_binary.
resource "cumulocity_application" "dashboard" {
  name         = "Fleet Dashboard"
  key          = "fleet-dashboard-key"
  type         = "HOSTED"
  context_path = "fleet-dashboard"
  availability = "PRIVATE"
  description  = "Custom dashboard for the field device fleet."
}

# An external application linking to a URL outside Cumulocity.
resource "cumulocity_application" "docs" {
  name = "Operations Handbook"
  key  = "ops-handbook-key"
  type = "EXTERNAL"
}

output "dashboard_id" {
  value = cumulocity_application.dashboard.id
}
