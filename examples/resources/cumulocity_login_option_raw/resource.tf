# The real OAuth2 client secret. The API masks it as "****" in responses, so it
# must be supplied here. Pass via TF_VAR_cognito_client_secret or a *.tfvars file.
variable "cognito_client_secret" {
  type      = string
  sensitive = true
}

# Full OAUTH2 (Cognito) login option managed as a verbatim JSON body.
#
# IMPORTANT: inside jsonencode() strings, `${...}` is Terraform interpolation.
# Cumulocity request-template placeholders must therefore be escaped as `$${...}`
# so Terraform emits a literal `${...}` to the API. Real interpolation (the
# client secret) uses a single `${...}`.
resource "cumulocity_login_option_raw" "cognito" {
  body = jsonencode({
    type                 = "OAUTH2"
    template             = "CUSTOM"
    providerName         = "cognito"
    buttonName           = "Login with Cognito"
    userManagementSource = "REMOTE"
    visibleOnLoginPage   = true
    grantType            = "AUTHORIZATION_CODE"
    useIdToken           = true

    issuer             = "https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_EXAMPLE1"
    audience           = "ClientID"
    clientId           = "ClientID"
    redirectToPlatform = "https://mytenant.cumulocity.com/tenant/oauth"

    userIdConfig = {
      useConstantValue = false
      jwtField         = "email"
      constantValue    = "email"
    }

    externalTokenConfig = {
      enabled = false
    }

    signatureVerificationConfig = {
      jwks = {
        jwksUri = "https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_EXAMPLE1/.well-known/jwks.json"
      }
    }

    accessTokenToUserDataMappings = {
      firstNameClaimName   = ""
      lastNameClaimName    = ""
      emailClaimName       = "email"
      phoneNumberClaimName = null
    }

    onNewUser = {
      dynamicMapping = {
        inventoryMappings = []
        configuration = {
          manageRolesOnlyFromAccessMapping = false
          mapRolesOnlyForNewUser           = false
          mapFromIdToken                   = false
        }
        mappings = [
          {
            id               = null
            thenApplications = []
            thenGroups       = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]
            when = {
              operator = "AND"
              childPredicates = [
                {
                  operator        = "IN"
                  parameterPath   = "cognito:groups"
                  value           = "C8Y_GLOBAL_ADMIN"
                  childPredicates = []
                }
              ]
            }
          }
        ]
      }
    }

    authorizationRequest = {
      method    = "GET"
      operation = "REDIRECT"
      url       = "https://example.auth.eu-central-1.amazoncognito.com/oauth2/authorize"
      headers   = {}
      requestParams = {
        response_type = "code"
        redirect_uri  = "$${redirectUri}"
        client_id     = "$${clientId}"
      }
    }

    tokenRequest = {
      method        = "POST"
      operation     = "EXECUTE"
      url           = "https://example.auth.eu-central-1.amazoncognito.com/oauth2/token"
      headers       = {}
      requestParams = {}
      body          = "grant_type=authorization_code&code=$${code}&redirect_uri=$${redirectUri}&client_id=$${clientId}&client_secret=${var.cognito_client_secret}"
    }

    refreshRequest = {
      method    = "POST"
      operation = "EXECUTE"
      url       = "https://example.auth.eu-central-1.amazoncognito.com/oauth2/token"
      headers   = {}
      body      = "grant_type=refresh_token&refresh_token=$${refreshToken}&client_id=$${clientId}"
      requestParams = {
        response_type = "refresh"
        client_secret = "${var.cognito_client_secret}"
      }
    }

    logoutRequest = {
      method    = "POST"
      operation = "REDIRECT"
      url       = "https://example.auth.eu-central-1.amazoncognito.com/logout"
      headers   = {}
      requestParams = {
        logout_uri = "https://mytenant.cumulocity.com"
        client_id  = "$${clientId}"
      }
    }
  })
}

# The server-side payload (with normalized/added fields and masked secrets).
output "cognito_id" {
  value = cumulocity_login_option_raw.cognito.id
}

output "cognito_effective_config" {
  value     = jsondecode(cumulocity_login_option_raw.cognito.config_json)
  sensitive = true
}
