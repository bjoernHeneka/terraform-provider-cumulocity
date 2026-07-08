
variable "tenant_domain" {
  type        = string
  description = "Cumulocity tenant domain, z.B. mytenant.cumulocity.com"
}

variable "tenant_id" {
  type        = string
  description = "Kurze Tenant-ID, z.B. t0071234 (optional)"
  default     = ""
}

variable "username" {
  type        = string
  description = "Login-Benutzername für den Provider"
}

variable "password" {
  type        = string
  sensitive   = true
  description = "Passwort für den Provider"
}
