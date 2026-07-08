# Lokales Test-Setup für den Cumulocity Terraform Provider.
# Kein "terraform init" nötig — dank dev_overrides in ~/.terraformrc
# wird der lokale Binary direkt verwendet.

terraform {
  required_providers {
    cumulocity = {
      source = "registry.terraform.io/bjoernHeneka/cumulocity"
    }
  }
}

provider "cumulocity" {
  # Werte können auch per Umgebungsvariable gesetzt werden:
  #   export CUMULOCITY_TENANT_DOMAIN=mytenant.cumulocity.com
  #   export CUMULOCITY_TENANT_ID=t0071234
  #   export CUMULOCITY_USERNAME=admin
  #   export CUMULOCITY_PASSWORD=secret

  tenant_domain = var.tenant_domain
  tenant_id     = var.tenant_id
  username      = var.username
  password      = var.password
}
