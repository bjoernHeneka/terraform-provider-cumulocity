# CLAUDE.md — Cumulocity Terraform Provider

## Project overview

This is a Terraform provider for the [Cumulocity IoT platform](https://www.cumulocity.com/api).
It uses **terraform-plugin-framework** (not the legacy SDKv2).

API reference (source of truth): the official [Cumulocity Core REST API documentation](https://cumulocity.com/api/core/), which also provides the OpenAPI specification. The spec is not vendored into this repository.

Go module: `github.com/bjoernHeneka/terraform-provider-cumulocity`
Provider address: `registry.terraform.io/bjoernHeneka/cumulocity`

## Directory structure

```
.
├── main.go                         # Provider entry point
├── go.mod                          # Go module file (run go mod tidy after changes)
├── GNUmakefile                     # build / test / install targets
├── internal/
│   ├── client/
│   │   └── client.go               # Typed HTTP client (Basic auth, doJSON helper)
│   └── provider/
│       ├── provider.go             # Provider schema, Configure(), Resources(), DataSources()
│       └── <resource>_resource.go  # One file per resource (add here as implemented)
└── examples/
    └── provider/
        └── provider.tf             # Provider usage example
```

## Authentication

Cumulocity uses **HTTP Basic auth** with a tenant-scoped credential format:

```
Authorization: Basic base64(<tenantID>/<username>:<password>)
```

The provider accepts four configuration attributes (all settable via env vars):

| Attribute       | Env var                    | Description                          |
|-----------------|----------------------------|--------------------------------------|
| `tenant_domain` | `CUMULOCITY_TENANT_DOMAIN` | e.g. `mytenant.cumulocity.com`       |
| `tenant_id`     | `CUMULOCITY_TENANT_ID`     | Short ID, e.g. `t0071234` (optional) |
| `username`      | `CUMULOCITY_USERNAME`      | Login username                       |
| `password`      | `CUMULOCITY_PASSWORD`      | Login password (sensitive)           |

The base URL is constructed as `https://<tenant_domain>`.

## HTTP client (`internal/client/client.go`)

- `NewClient(tenantDomain, tenantID, username, password)` — validates and builds the client
- `doJSON(ctx, method, path, input, out)` — generic request helper; returns `ErrNotFound` on 404
- `GetCurrentUser(ctx)` — validates credentials; useful for provider acceptance tests

When adding new API operations, add a method to `Client` that calls `c.doJSON(...)`.
Add request/response structs in the same file or a new file in `internal/client/`.

## Adding a new resource

1. Create `internal/provider/<name>_resource.go` following the full CRUD pattern
2. Add `client.<Resource>` types and CRUD methods to `internal/client/client.go`
3. Register the constructor in `provider.go`: `Resources()` → append `New<Name>Resource`
4. Add an example in `examples/resources/cumulocity_<name>/resource.tf`

### Non-negotiable resource rules

- Every resource model **must** have `ID types.String \`tfsdk:"id"\`` with `Computed: true`
- Model fields use `types.String`, `types.Int64`, `types.Bool` — never plain Go types
- `Read` must call `resp.State.RemoveResource(ctx)` when `errors.Is(err, client.ErrNotFound)`
- Sensitive fields (credentials, tokens) need `Sensitive: true` in the schema
- Every resource implements `ImportState` via `resource.ImportStatePassthroughID`

## Available API resources

Key resources to implement (full CRUD available):

| Resource                  | Base path                          |
|---------------------------|------------------------------------|
| Managed Object (inventory)| `/inventory/managedObjects`        |
| Alarm                     | `/alarm/alarms`                    |
| Event                     | `/event/events`                    |
| Operation (device control)| `/devicecontrol/operations`        |
| Application               | `/application/applications`        |
| User                      | `/user/{tenantId}/users`           |
| Tenant                    | `/tenant/tenants`                  |
| Retention Rule            | `/retention/retentions`            |
| Notification Subscription | `/notification2/subscriptions`     |
| Trusted Certificate       | `/tenant/trusted-certificates`     |
| Login Option              | `/tenant/loginOptions`             |

## Build & test

```bash
# Build
go build ./...

# Unit tests
go test ./internal/...

# Acceptance tests (requires real Cumulocity credentials)
export CUMULOCITY_TENANT_DOMAIN=mytenant.cumulocity.com
export CUMULOCITY_TENANT_ID=t0071234
export CUMULOCITY_USERNAME=admin
export CUMULOCITY_PASSWORD=secret
TF_ACC=1 go test -v ./internal/provider/...

# Install locally for manual testing
make install
```

## Local dev override (`~/.terraformrc`)

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/bjoernHeneka/cumulocity" = "/path/to/your/bin"
  }
  direct {}
}
```

## Code conventions

- Use `fmt.Errorf("context: %w", err)` for error wrapping
- No bare `panic()` calls in provider code
- Keep resource files focused: one resource per file
