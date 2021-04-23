# Release Process

1. Merge PR.
2. Run full acceptance test suite (to be sure all is well before 'cutting a new release').
3. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/348)).
4. Rebase latest remote main branch locally (`git pull --rebase origin main`).
5. Tag a new release (`git tag -s vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`).
6. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
7. Publish draft release.
8. Communicate the release in the relevant Slack channels.

## Notes

Step 3. we utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.

Step 5. causes a [github action](https://github.com/fastly/terraform-provider-fastly/blob/main/.github/workflows/release.yml) to be triggered which produces a 'draft' release.

Step 7. causes a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) to be triggered which produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest) and which can take a while to publish, so check back later if it doesn't show up within a few minutes.

Step 8. `#api-clients`, `#ecp-languages`.
