# Upload a firmware image to the Cumulocity inventory binary store.
resource "cumulocity_binary" "firmware" {
  file         = "${path.module}/firmware-1.0.0.bin"
  file_hash    = filemd5("${path.module}/firmware-1.0.0.bin")
  name         = "firmware-1.0.0.bin"
  content_type = "application/octet-stream"
}

output "firmware_id" {
  value = cumulocity_binary.firmware.id
}
