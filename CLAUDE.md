# terraform-provider-fastly

A Terraform provider for the [Fastly](https://www.fastly.com/) API, built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

Full design spec: `docs/dual-model-provider-design.md`. Query support design: `docs/terraform-query.md`.

## Core architecture

Two resource families encoded in the resource type name (never a `mode`/`type` field):

- **Explicit** (`fastly_service_cdn`, `fastly_service_compute`, `fastly_service_domain`, `fastly_service_backend`) — caller manages version lifecycle. Provider writes to the version specified; rejects writes to active/locked versions.  
- **Automatic** (`fastly_service_cdn_auto`, `fastly_service_compute_auto`) — provider auto-clones, validates, and activates every CRUD. Nested blocks own versioned config. No staging support.

`terraform query` and actions (`version_clone`, `version_activate`, `version_stage`, `compute_package_upload`) are explicit-only.

## Project structure

- `internal/resources/<name>/` — one package per resource (`resource.go`, `schema.go`, `expand.go`, `flatten.go`, `list.go`)  
- `internal/actions/<name>/` — provider actions (explicit-only)  
- `internal/datasources/<name>/` — data sources (e.g. `serviceversion/`)  
- `internal/service/` — version selection, mutability gating, service-type helpers  
- `internal/provider/` — provider registration (`provider.go`)  
- `internal/acceptance_tests/` — acceptance tests (`TF_ACC=1`), config builders, and fixtures  
- `examples/` — working HCL examples per workflow pattern  
- `docs/` — Registry documentation: design specs at the top level, plus generated per-object docs under `docs/resources/`, `docs/data-sources/`, `docs/actions/`, `docs/list-resources/`  
- `templates/` — `tfplugindocs` templates that drive `docs/` generation (edit these, not the generated `.md` files)

## Dev workflow

### Building

```shell
make build   # runs go fmt, compiles binary, writes bin/developer_overrides.tfrc
```

### Running locally

```shell
export TF_CLI_CONFIG_FILE=$(pwd)/bin/developer_overrides.tfrc
export FASTLY_API_TOKEN=<token>
cd examples/<dir> && terraform apply
```

### Formatting

```shell
make fmt     # go fmt ./...
```

### Testing

- `make test-unit` — no API token needed; unit tests live next to the code (e.g. `backend/backend_test.go`)
- `make test-acc` / `make test-lifecycle` — acceptance + full apply/destroy lifecycle; both require `FASTLY_API_TOKEN`

Acceptance tests live in `internal/acceptance_tests/` and build HCL via `config_builder.go`.

### Documentation

Per-object docs in `docs/` are **generated** by `tfplugindocs` — do not hand-edit the generated `.md` files. Edit schema descriptions in code and the templates under `templates/`, then run `make docs`.

## Code style

- Go 1.26; formatted with `gofmt` (enforced via `make fmt` / `make build`)  
- No comments explaining *what* code does — only *why* when non-obvious  
- `StringValue`, `Int64Value`, `BoolValue` unwrap `types.*` safely; prefer these over direct field access  
- `ToGeneratedResourceName(parts...)` builds HCL identifiers from API names (used in list resources)  
- Use `new(T)` to take a pointer to a value — prefer this over `fastly.ToPointer` (a go-fastly helper that has been replaced with idiomatic Go)

## Conventions

- New resource packages live under `internal/resources/<name>/`; actions under `internal/actions/<name>/`. See "Adding a new resource" below for the full checklist.  
- **Canonical template for a versioned explicit resource:** `internal/resources/backend/` — it is the most complete example (all five files, identity schema, import, list resource, NestedModel).  
- **Canonical template for a service resource:** `internal/resources/servicecdn/` (explicit) or `internal/resources/servicecdnauto/` (automatic).  
- **NestedModel pattern:** if a versioned resource should also be usable inside `_auto` aggregate resources, the package must additionally export `NestedModel`, `NestedBlockSchema()`, `ReadForVersion()`, `Reconcile()`, `Equal()`, and `ModelsEqual()` — see `internal/resources/backend/schema.go`. Not all versioned resources need this; only those that are nested in an auto resource do.  
- Register in `provider.go`: `Resources()`, `ListResources()`, `DataSources()`, or `Actions()`.  
- Service-kind validation: use `service.TypeSupported()` \+ a clear diagnostic. Cannot rely on `terraform validate` because `service_id` may be computed.  
- Read/import/query paths: always use `service.SelectReadVersion()` — never mutates. Mutability gating: `VersionChecker.GetMutability()`.  
- `service.TypeVCL = "vcl"`, `service.TypeCompute = "wasm"`.

## Adding a new resource

End-to-end checklist for a standalone resource (use `internal/resources/backend/` as the template):

1. Create `internal/resources/<name>/` with `resource.go`, `schema.go`, `expand.go`, `flatten.go`, `list.go` (omit files with no content).
2. If it nests inside an `_auto` resource, add the `NestedModel` exports — see the NestedModel pattern under Conventions.
3. Register in `provider.go` (`Resources()`, and `ListResources()` if it has a list resource).
4. Add an acceptance test under `internal/acceptance_tests/` and HCL under `examples/`.
5. Run `make docs` to generate the Registry documentation.

## Adding a nested block to an auto resource

When a new versioned resource should appear as a nested block inside `_auto` service resources, there are four touchpoints in each auto resource file (`servicecdnauto/resource.go`, `servicecomputeauto/resource.go`):

1. Add `[]<pkg>.NestedModel` field to `Model` struct  
2. Add `<pkg>.NestedBlockSchema()` to `Schema` Blocks map  
3. Call `<pkg>.Reconcile(...)` in Create and Update  
4. Call `<pkg>.ReadForVersion(...)` in Read, assign result to state

See `internal/resources/servicecdnauto/resource.go` for the exact pattern.

## Dependencies

- `github.com/fastly/go-fastly/v15` — Fastly API client  
- `github.com/hashicorp/terraform-plugin-framework v1.17.0` — includes `action` and `list` packages

## Related

- The existing (SDKv2) provider this rewrite replaces lives on [`main`](https://github.com/fastly/terraform-provider-fastly/tree/main) — reference it for feature and behavior parity.

