# Release Process

1. Merge PR.
2. Run full acceptance test suite (to be sure all is well before 'cutting a new release').
3. Open a new PR to update CHANGELOG ([example](https://github.com/fastly/terraform-provider-fastly/pull/348)).
4. Rebase latest remote master branch locally (`git pull --rebase origin master`).
5. Tag a new release (`git tag -s vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`).
6. Copy/paste CHANGELOG into the [draft release](https://github.com/fastly/terraform-provider-fastly/releases).
7. Publish draft release.

## Notes

Step 5. causes a [github action](https://github.com/fastly/terraform-provider-fastly/blob/master/.github/workflows/release.yml) to be triggered which produces a 'draft' release.

Step 7. causes a [github webhook](https://github.com/fastly/terraform-provider-fastly/settings/hooks) to be triggered which produces a release on the [terraform registry](https://registry.terraform.io/providers/fastly/fastly/latest) and which can take a while to publish, so check back later if it doesn't show up within a few minutes.
