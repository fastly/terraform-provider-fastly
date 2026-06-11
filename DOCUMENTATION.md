# Provider Documentation

The provider documentation is published by the Terraform Registry from the `docs/`
directory in this repository. Documentation is versioned with provider releases,
so any documentation-only change requires a new provider release before it appears
in the Registry.

The dual-model provider uses the Terraform Registry documentation layout:

```text
docs/
  index.md
  guides/
  resources/
  data-sources/
  actions/
  list-resources/
```

Resource, action, data source, and list-resource documentation filenames should
not include the `fastly_` provider prefix. For example:

```text
docs/resources/service_cdn.md
docs/actions/service_version_activate.md
docs/data-sources/service_version.md
docs/list-resources/service_backend.md
```

## Commands

Generate documentation from provider schemas and documentation templates:

```shell
make generate-docs
```

Validate the documentation structure and provider schema references:

```shell
make validate-docs
```

Generate and validate in one step:

```shell
make docs
```

Preview generated Markdown by copying a generated file into the Terraform
Registry documentation preview tool:

https://registry.terraform.io/tools/doc-preview

## Pull request checks

The pull request workflow runs documentation checks in a dedicated `docs` job.

The job runs unless the pull request has the `Skip-Docs` label. It performs the
same commands that should be run locally:

```shell
make generate-docs
make validate-docs
git diff --exit-code --ignore-all-space ./docs/
```

The final diff check ensures generated documentation has been committed. If the
job fails because `./docs/` changed, run `make generate-docs`, review the
generated changes, and commit the updated documentation.

## Adding or changing provider schema

When adding a new resource, data source, action, list resource, function, or
ephemeral resource:

1. Register it in the provider implementation.
2. Add descriptions to every schema attribute and nested block.
3. Add an example to the documentation page.
4. Add import documentation for managed resources where import is supported.
5. Run `make generate-docs`.
6. Run `make validate-docs`.
7. Review the generated Markdown before opening the pull request.

The documentation generator reads the provider schema from the provider itself.
If a resource, data source, action, or list resource is not registered in the
provider, generated documentation for it will be incomplete or absent.

## Dual-model documentation expectations

The documentation must clearly separate the two service lifecycle families.

### Automatic compatibility resources

Automatic resources use the `_auto` suffix, such as:

- `fastly_service_cdn_auto`
- `fastly_service_compute_auto`

These resources own the service version lifecycle. They select or create a
writable version, reconcile nested configuration, validate the version, and
activate it.

### Explicit/default resources

Explicit/default resources use clean resource names, such as:

- `fastly_service_cdn`
- `fastly_service_compute`
- `fastly_service_domain`
- `fastly_service_backend`

These resources do not clone, activate, or stage service versions during normal
CRUD. Users manage service versions explicitly with actions such as
`fastly_service_version_clone`, `fastly_service_version_activate`, and
`fastly_service_version_stage`.

## Documentation checklist

Before merging documentation changes, confirm that:

- `docs/index.md` exists.
- Each registered resource has a page in `docs/resources/`.
- Each registered data source has a page in `docs/data-sources/`.
- Each registered action has a page in `docs/actions/`.
- Each registered list resource has a page in `docs/list-resources/`.
- User-facing examples are valid HCL.
- Managed resources document import behavior where supported.
- Automatic and explicit lifecycle behavior are not mixed.
- `make generate-docs` has been run.
- `make validate-docs` passes.
- `git diff --exit-code --ignore-all-space ./docs/` passes after generation.
