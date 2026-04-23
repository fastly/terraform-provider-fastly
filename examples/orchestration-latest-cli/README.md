# Orchestration Example (Latest-Version Convenience Workflow, CLI)

This example demonstrates a convenience-oriented workflow for the version-agnostic
Fastly Terraform demo provider. It targets the latest version for each service
and keeps version lifecycle explicit, but uses the Fastly CLI for clone and
activate operations.

This variant improves the original latest-version example by using one helper
script to clone only the services that actually changed, and one helper script
to activate only the services that changed.

## One-time bootstrap

```bash
cd examples/orchestration-latest-cli
terraform init
terraform apply
```

## Day-to-day engineer workflow

```bash
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform show -json tfplan > tfplan.json
./scripts/clone-changed-services.sh tfplan.json
terraform apply
./scripts/activate-changed-services.sh tfplan.json
```

## Notes

- This is a convenience workflow, not the recommended CI/CD workflow.
- Because the configuration targets `latest_version`, this is best suited to
  careful single-user workflows.
- The key improvement over the original example is that cloning now mirrors
  activation: both operate only on the changed services.
