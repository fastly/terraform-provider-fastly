4# Release Process

1. Merge all PRs intended for the release.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`).
1. Ensure all analysis checks and tests are passing (`TEST_PARALLELISM=8 make testacc`).
1. Run `go mod vendor` and `make goreleaser GORELEASER_ARGS="--snapshot --skip=validate --clean"`.
1. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/498/files))<sup>[1](#note1)</sup>.
1. 🚨 Ensure any _removals_ are considered a BREAKING CHANGE and must be published in a major release.
1. Merge CHANGELOG.
1. Rebase latest remote main branch locally (`git pull --rebase origin main`)<sup>[2](#note2)</sup>.
1. Tag a new release (`tag=vX.Y.Z && git tag -s $tag -m $tag && git push $(git config branch.$(git symbolic-ref -q --short HEAD).remote) $tag`)<sup>[3](#note3)</sup>.
1. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
1. Publish draft release<sup>[4](#note4)</sup>.

## Footnotes

1. <a name="note1"></a>We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.
1. <a name="note2"></a>🚨 Manually update generated `docs/index.md` and force push (as we're not able to update the git tag until the next step).
1. <a name="note3"></a>Triggers a [github action](https://github.com/fastly/terraform-provider-fastly/blob/main/.github/workflows/release.yml) that produces a 'draft' release.
1. <a name="note4"></a>Triggers a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) that produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest).
