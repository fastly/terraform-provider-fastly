# Release Process

1. Merge all PRs intended for the release.
2. Rebase latest remote main branch locally (`git pull --rebase origin main`).
3. Ensure all analysis checks and tests are passing (`TEST_PARALLELISM=8 make testacc`).
4. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/348))<sup>[1](#note1)</sup>.
5. Merge CHANGELOG.
6. Rebase latest remote main branch locally (`git pull --rebase origin main`).
7. Tag a new release (`git tag -s vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`)<sup>[2](#note2)</sup>.
8. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
9. Publish draft release<sup>[3](#note3)</sup>.
10. Communicate the release in the relevant Slack channels<sup>[4](#note4)</sup>.

## Footnotes

1. <a name="note1"></a>We utilize [semantic versioning](https://semver.org/) and only include relevant/significant changes within the CHANGELOG.
2. <a name="note2"></a>Triggers a [github action](https://github.com/fastly/terraform-provider-fastly/blob/main/.github/workflows/release.yml) that produces a 'draft' release.
3. <a name="note3"></a>Triggers a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) that produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest).
4. <a name="note4"></a>Fastly make internal announcements in the Slack channels: `#api-clients`, `#ecp-languages`.
