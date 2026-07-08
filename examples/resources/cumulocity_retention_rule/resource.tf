# Keep all alarms for 30 days.
resource "cumulocity_retention_rule" "alarms" {
  data_type   = "ALARM"
  maximum_age = 30
}

# Keep temperature measurements for 90 days.
resource "cumulocity_retention_rule" "temperature" {
  data_type     = "MEASUREMENT"
  fragment_type = "c8y_TemperatureMeasurement"
  maximum_age   = 90
}

# Keep everything (catch-all) for 365 days.
resource "cumulocity_retention_rule" "default" {
  data_type   = "*"
  maximum_age = 365
}
