resource "cumulocity_application" "dashboard" {
  name         = "Fleet Dashboard"
  key          = "fleet-dashboard-key"
  type         = "HOSTED"
  context_path = "fleet-dashboard"
}

# Upload the ZIP archive for the application. Each upload creates a new binary
# version and activates it. file_hash triggers a re-upload when the archive
# content changes but the path stays the same.
resource "cumulocity_application_binary" "dashboard_bundle" {
  application_id = cumulocity_application.dashboard.id
  file           = "${path.module}/dashboard.zip"
  file_hash      = filemd5("${path.module}/dashboard.zip")
}

output "active_binary_id" {
  value = cumulocity_application_binary.dashboard_bundle.binary_id
}
