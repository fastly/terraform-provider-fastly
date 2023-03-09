# Release Process

1. Merge all PRs intended for the release.
2. Rebase latest remote main branch locally (`git pull --rebase origin main`).
3. Ensure all analysis checks and tests are passing (`TEST_PARALLELISM=8 make testacc`).
4. Run `go mod vendor` and `make goreleaser GORELEASER_ARGS="--skip-validate --rm-dist"`.
5. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/498/files))<sup>[1](#note1)</sup>.
6. Merge CHANGELOG.
7. Rebase latest remote main branch locally (`git pull --rebase origin main`)<sup>[2](#note2)</sup>.
8. Tag a new release (`tag=vX.Y.Z && git tag -s $tag -m "$tag" && git push origin $tag`)<sup>[3](#note3)</sup>.
9. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
10. Publish draft release<sup>[4](#note4)</sup>.
11. Communicate the release in the relevant Slack channels<sup>[5](#note5)</sup>.

## Footnotes

1. <a name="note1"></a>We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.
2. <a name="note2"></a>Manually update generated `docs/index.md` and force push (as we're not able to update the git tag until the next step).
3. <a name="note3"></a>Triggers a [github action](https://github.com/fastly/terraform-provider-fastly/blob/main/.github/workflows/release.yml) that produces a 'draft' release.
4. <a name="note4"></a>Triggers a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) that produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest).
5. <a name="note5"></a>Fastly make internal announcements in the Slack channels: `#api-clients`, `#ecp-languages`.
