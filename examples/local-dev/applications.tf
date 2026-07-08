
# A hosted web application
resource "cumulocity_application" "webui" {
  key          = "tf-test-webui-key"
  name         = "tf-test-webui"
  type         = "HOSTED"
  context_path = "tf-test-webui"
  availability = "PRIVATE"
  description  = "Test web application managed by Terraform"
}

# Upload the application ZIP — re-upload when file content changes via file_hash
# resource "cumulocity_application_binary" "webui_zip" {
#   application_id = cumulocity_application.webui.id
#   file           = "${path.module}/myapp.zip"
#   file_hash      = filemd5("${path.module}/myapp.zip")
# }

output "webui_id" {
  value       = cumulocity_application.webui.id
  description = "Application ID — use with cumulocity_application_binary.application_id"
}

output "webui_active_version" {
  value       = cumulocity_application.webui.active_version_id
  description = "Active binary version ID (set after cumulocity_application_binary upload)"
}
