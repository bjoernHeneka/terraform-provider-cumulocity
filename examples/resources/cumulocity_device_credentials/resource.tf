# Request auto-generated credentials for a device during the bootstrap flow.
#
# NOTE: POST /devicecontrol/deviceCredentials is restricted to the Cumulocity
# bootstrap user (devicebootstrap). Configure the provider with bootstrap
# credentials to use this resource; regular admin credentials return HTTP 403.
resource "cumulocity_device_credentials" "gateway" {
  device_id      = "SN-00123456"
  security_token = var.device_security_token
}

variable "device_security_token" {
  type      = string
  sensitive = true
}

# The generated password is only returned on creation and stored in state.
output "device_username" {
  value = cumulocity_device_credentials.gateway.username
}

output "device_password" {
  value     = cumulocity_device_credentials.gateway.password
  sensitive = true
}
