# Release Process

## Prerequisites

For security we sign tags. To be able to sign tags you need to tell Git which key you would like to use. Please follow these
[steps](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key) to
tell Git about your signing key.

## Steps

1. Merge all PRs intended for the release.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Ensure all analysis checks and tests are passing (`TEST_PARALLELISM=8 make testacc`).
1. Run `make goreleaser GORELEASER_ARGS="--snapshot --skip=validate --clean"`.
1. Manually update generated `docs/index.md`.
1. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/498/files)).
    - make sure to use the `Skip-Docs` label before opening to ensure the docs action doesn't fail with the new version.
    - We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.
1. 🚨 Ensure any _removals_ are considered a BREAKING CHANGE and must be published in a major release.
1. Merge CHANGELOG.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Create a new signed tag (replace `{{remote}}` with the remote pointing to the official repository i.e. `origin` or `upstream` depending on your Git workflow): `tag=vX.Y.Z && git tag -s $tag -m $tag && git push {{remote}} $tag`.
    - Triggers a [github action](https://github.com/fastly/terraform-provider-fastly/blob/main/.github/workflows/release.yml) that produces a 'draft' release.
1. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
1. Publish draft release.
    - Triggers a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) that produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest).
