# Cumulocity Terraform Provider — Implementierungsstand

Letzte Aktualisierung: 2026-07-02

Dieses Dokument listet die tatsächlich implementierten Ressourcen und Data Sources
auf und benennt offen die bekannten Lücken gegenüber der Cumulocity-API. Die Zahlen
beziehen sich ausschließlich auf das, was der Provider heute registriert
(`internal/provider/provider.go`) — nicht auf selbstgesteckte Teilziele.

## Ressourcen (26)

| API-Gruppe | Cumulocity Endpunkt | Terraform Ressource |
|---|---|---|
| **Alarm API** | `/alarm/alarms` | `cumulocity_alarm` |
| **Audit API** | `/audit/auditRecords` | `cumulocity_audit_record` |
| **Device Control API** | `/devicecontrol/operations` | `cumulocity_device_operation` |
| **Device Control API** | `/devicecontrol/newDeviceRequests` | `cumulocity_new_device_request` |
| **Device Control API** | `/devicecontrol/deviceCredentials` | `cumulocity_device_credentials` |
| **Device Control API** | `/devicecontrol/bulkoperations` | `cumulocity_bulk_operation` |
| **Event API** | `/event/events` | `cumulocity_event` |
| **Application API** | `/application/applications` | `cumulocity_application` |
| **Application API** | `/application/applications/{id}/binaries` | `cumulocity_application_binary` |
| **Identity API** | `/identity/externalIds` | `cumulocity_external_id` |
| **Inventory API** | `/inventory/managedObjects` | `cumulocity_managed_object` |
| **Inventory API** | `/inventory/binaries` | `cumulocity_binary` |
| **Measurement API** | `/measurement/measurements` | `cumulocity_measurement` |
| **Notification 2.0 API** | `/notification2/subscriptions` | `cumulocity_notification_subscription` |
| **Retention API** | `/retention/retentions` | `cumulocity_retention_rule` |
| **Tenant API** | `/tenant/tenants` | `cumulocity_tenant` |
| **Tenant API** | `/tenant/tenants/{tenantId}/applications` | `cumulocity_tenant_application_subscription` |
| **Tenant API** | `/tenant/tenants/{tenantId}/trusted-certificates` | `cumulocity_trusted_certificate` |
| **Tenant API** | `/tenant/loginOptions` | `cumulocity_login_option` |
| **Tenant API** | `/tenant/loginOptions` (Roh-JSON) | `cumulocity_login_option_raw` |
| **Tenant API** | `/tenant/options` | `cumulocity_tenant_option` |
| **User API** | `/user/{tenantId}/users` | `cumulocity_user` |
| **User API** | `/user/{tenantId}/users/{id}/roles` | `cumulocity_user_role_assignment` |
| **User API** | `/user/{tenantId}/users/{id}/roles/inventory` | `cumulocity_user_inventory_role_assignment` |
| **User API** | `/user/{tenantId}/groups` | `cumulocity_user_group` |
| **User API** | `/user/{tenantId}/groups/{id}/users` | `cumulocity_user_group_membership` |

## Data Sources (15)

| API-Gruppe | Cumulocity Endpunkt | Terraform Data Source |
|---|---|---|
| **User API** | `/user/roles/{name}` | `cumulocity_role` |
| **User API** | `/user/roles` | `cumulocity_roles` |
| **User API** | `/user/inventoryroles` | `cumulocity_inventory_role` |
| **User API** | `/user/inventoryroles` | `cumulocity_inventory_roles` |
| **Tenant API** | `/tenant/options` | `cumulocity_tenant_options` |
| **Tenant API** | `/tenant/loginOptions/{typeOrId}` | `cumulocity_login_option` |
| **Tenant API** | `/tenant/loginOptions` | `cumulocity_login_options` |
| **Device Control API** | `/devicecontrol/operations` | `cumulocity_operations` |
| **Alarm API** | `/alarm/alarms` | `cumulocity_alarms` |
| **Audit API** | `/audit/auditRecords` | `cumulocity_audit_records` |
| **Event API** | `/event/events` | `cumulocity_events` |
| **Inventory API** | `/inventory/managedObjects` | `cumulocity_managed_objects` |
| **Inventory API** | `/inventory/binaries` | `cumulocity_binaries` |
| **Measurement API** | `/measurement/measurements` | `cumulocity_measurements` |
| **Application API** | `/application/applicationsByName/{name}` | `cumulocity_application` |

## Zusammenfassung

| Kategorie | Implementiert |
|---|:---:|
| Ressourcen | 26 |
| Data Sources | 15 |

Die Cumulocity-API ist deutlich größer als dieser Umfang. Die folgenden Abschnitte
benennen die relevanten Lücken. "Vollständig" ist der Provider damit ausdrücklich
**nicht** — er deckt die gängigsten Provisioning-Anwendungsfälle ab.

## Fehlende Ressourcen

### Terraform-tauglich, sinnvoll, aber noch nicht vorhanden

- **`inventory_role` als Ressource** — `/user/inventoryroles` (POST/PUT/DELETE).
  Aktuell nur als Data Source lesbar; das Anlegen und Verwalten eigener Inventory-Rollen
  fehlt.
- **`group_role_assignment`** — `/user/{tenantId}/groups/{groupId}/roles`.
  Zuweisung globaler Rollen an Benutzergruppen.
- **Beziehungs-Ressourcen `child_device` / `child_asset` / `child_addition`** —
  `/inventory/managedObjects/{id}/childDevices`, `.../childAssets`, `.../childAdditions`.
  Modellierung der Inventar-Hierarchie (Gerät ↔ Gruppe/Asset).
- **SSO `access_mapping` + `inventory_access_mapping`** —
  `/tenant/loginOptions/{typeOrId}/accessMappings` bzw. `.../inventoryAccessMappings`.
  Mapping von SSO-Claims auf Rollen bzw. Inventory-Zuweisungen.
- **`application_version`** — `/application/applications/{id}/versions`.
  Verwaltung mehrerer Versionen einer Anwendung.
- **`feature_toggle` pro Tenant** — `/features/{featureKey}/by-tenant/{tenantId}`.
  Setzen von Feature-Flags je Tenant (nicht die read-only Gesamtliste).

### Grenzfälle (Nutzen abhängig vom Anwendungsfall)

- **Event-Binaries** — Datei-Anhänge an Events.
- **`devicePermissions`** — feingranulare Geräteberechtigungen.
- **Login-Option `restrict`** — eher ein Attribut bestehender Login-Optionen als eigene Ressource.
- **Tenant-TFA** — Zwei-Faktor-Einstellungen je Tenant.

## Fehlende Data Sources zu bereits existierenden Ressourcen

Für mehrere Ressourcen fehlt bislang das passende Lookup/Listing:

- `users`
- `tenants`
- `applications` (Liste)
- `user_groups`
- `trusted_certificates`
- `retention_rules`
- `notification_subscriptions`
- `external_id` (Lookup)
- `new_device_requests`
- `bulk_operations`
- `tenant_option` (Single-Lookup; aktuell nur die Liste `tenant_options`)

## Bewusst nicht abgedeckt (read-only / Laufzeit / Streaming)

Die folgenden API-Bereiche sind als Terraform-Ressourcen nicht sinnvoll, da sie
reine Lese-Endpunkte, Laufzeit- oder Streaming-Funktionen sind bzw. systemseitig
verwaltet werden:

- Platform-API (`/platform`) — Systeminfo, read-only
- Current Tenant (`/tenant/currentTenant`) — Laufzeitinformation
- Usage Statistics (`/tenant/statistics/...`) — read-only
- System Options (`/tenant/system/options`) — nur lesbar
- OAuth/Token-Endpunkte — Laufzeit-Tokens
- Current Application (`/application/currentApplication/...`) — nur zur Laufzeit relevant
- Current User (`/user/currentUser`) — Laufzeitinformation
- Roles (`/user/roles`) — nur GET, Data Source vorhanden
- Measurement Series (`/measurement/measurements/series`) — Aggregat-Abfrage
- Alarm Count (`/alarm/alarms/count`) — Aggregat-Abfrage
- Notification 2.0 Token (`/notification2/token`) — kurzlebige Tokens
- Realtime Notifications (`/notification/realtime`) — WebSocket-basiert
- Device Access Token — Laufzeit-Token
- Certificate Authority / EST (`/certificate-authority`, EST-Endpunkte) — spezialisierter Fluss
- Verify-Cert-Chain — Verifikationsaufruf, kein verwalteter Zustand
- HATEOAS-Discovery-Roots — reine API-Navigationseinstiege
