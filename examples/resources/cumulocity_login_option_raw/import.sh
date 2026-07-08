# Import an existing login option by its ID (or type).
# Find the ID via the data source or GET /tenant/loginOptions.
terraform import cumulocity_login_option_raw.cognito 10a66775-4819-436f-8d43-f17d0409cea2

# After import:
#   1. `terraform show`  -> read the config_json output to see the current config.
#   2. Author the `body` argument to match, replacing any masked "****" secrets
#      with the real values (e.g. via var.cognito_client_secret).
#   3. `terraform plan`  -> the first plan after import will show an update
#      because `body` is not populated by import; applying it re-asserts your
#      declared configuration.
