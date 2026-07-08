---
page_title: "Data Source: cumulocity_managed_objects"
description: |-
  Lists Cumulocity managed objects, with optional type, fragment, query, text, and owner filters.
---

# cumulocity_managed_objects

Retrieves a list of managed objects (devices, assets, groups, etc.) from the Cumulocity inventory, optionally filtered by type, fragment, inventory query, text, or owner. All pages are followed automatically.

Corresponds to `GET /inventory/managedObjects`.

## Example Usage

```hcl
# All devices of a given type
data "cumulocity_managed_objects" "mqtt_devices" {
  fragment_type = "c8y_IsDevice"
  type          = "c8y_MQTTDevice"
}

output "device_names" {
  value = [for mo in data.cumulocity_managed_objects.mqtt_devices.managed_objects : mo.name]
}
```

## Schema

### Optional

- `type` (String) — Filter by managed object type, e.g. `c8y_MQTTDevice`.
- `fragment_type` (String) — Filter by the presence of a specific fragment, e.g. `c8y_IsDevice`.
- `query` (String) — Advanced inventory query string (Cumulocity query language).
- `text` (String) — Full-text search — returns objects whose name contains this string.
- `owner` (String) — Filter by owner username.

### Read-Only

- `managed_objects` (List of Object) — List of matching managed objects. Each object contains:
  - `id` (String) — Managed object ID.
  - `name` (String) — Name of the managed object.
  - `type` (String) — Type of the managed object.
  - `owner` (String) — Owner username.
  - `self` (String) — Self-link URL.
  - `creation_time` (String) — ISO 8601 creation timestamp.
  - `last_updated` (String) — ISO 8601 last update timestamp.
  - `is_device` (Boolean) — Whether the object carries `c8y_IsDevice`.
  - `is_device_group` (Boolean) — Whether the object carries `c8y_IsDeviceGroup`.
