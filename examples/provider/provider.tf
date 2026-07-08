
terraform {
  required_providers {
    cumulocity = {
      source  = "registry.terraform.io/org-codebee/cumulocity"
      version = "~> 0.1"
    }
  }
}

# Option 1: explicit configuration
provider "cumulocity" {
  # Tenant domain, e.g. "mytenant.cumulocity.com"
  tenant_domain = "mytenant.cumulocity.com"

  # Short tenant ID used in Basic auth (e.g. "t0071234")
  # If your platform does not require a tenant prefix, omit this.
  tenant_id = "t0071234"

  # Username and password for Basic auth
  username = "admin"
  password = var.cumulocity_password
}

# Option 2: environment variables (recommended for CI/CD)
# Set the following before running terraform:
#   export CUMULOCITY_TENANT_DOMAIN=mytenant.cumulocity.com
#   export CUMULOCITY_TENANT_ID=t0071234
#   export CUMULOCITY_USERNAME=admin
#   export CUMULOCITY_PASSWORD=secret
#
# Then use an empty provider block:
# provider "cumulocity" {}

variable "cumulocity_password" {
  type      = string
  sensitive = true
}
