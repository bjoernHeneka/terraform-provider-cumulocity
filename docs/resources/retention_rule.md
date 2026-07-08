---
page_title: "Resource: cumulocity_retention_rule"
description: |-
  Manages a Cumulocity retention rule.
---

# cumulocity_retention_rule

Manages a Cumulocity retention rule, which controls how long data of a given type is kept.

Corresponds to `POST/GET/PUT/DELETE /retention/retentions/{id}`.

## Example Usage

```hcl
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
```

## Schema

### Required

- `maximum_age` (Number) — Maximum age of matching data, expressed in number of days.

### Optional

- `data_type` (String) — The data type(s) to which the rule applies. One of: `ALARM`, `AUDIT`, `BULK_OPERATION`, `EVENT`, `MEASUREMENT`, `OPERATION`, `*`. Defaults to `*`.
- `fragment_type` (String) — The fragment type(s) to which the rule applies. Used by `EVENT`, `MEASUREMENT`, `OPERATION` and `BULK_OPERATION`. Defaults to `*`.
- `source` (String) — The source(s) to which the rule applies. Defaults to `*`.
- `type` (String) — The type(s) to which the rule applies. Used by `ALARM`, `AUDIT`, `EVENT` and `MEASUREMENT`. Defaults to `*`.

### Read-Only

- `id` (String) — Unique identifier assigned by Cumulocity.
- `editable` (Boolean) — Whether the rule is editable. Set to `false` by the platform for system-managed rules; can only be changed by the Management tenant.
- `self` (String) — Self-link URL of the retention rule.

## Import

Import an existing retention rule by its ID:

```shell
terraform import cumulocity_retention_rule.alarms 20200301
```
