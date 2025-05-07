# Contributing

We're happy to receive feature requests and PRs. If your change is nontrivial,
please open an [issue](https://github.com/fastly/terraform-provider-fastly/issues/new) to discuss the
idea and implementation strategy before submitting a PR.

1. Fork the repository.
    1. Please also following the instructions for: [Allowing changes to a pull request branch created from a fork](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork)
2. Create an `upstream` remote.
```bash
$ git remote add upstream git@github.com:fastly/terraform-provider-fastly.git
```
3. Create a feature branch.
4. Make changes.
5. Write tests.
6. Validate your change via the steps documented [in the README](./README.md#testing).
7. Open a pull request against `upstream main`.
    1. Once you have marked your PR as `Ready for Review` please do not force push to the branch
8. Add an entry in `CHANGELOG.md` in the `UNRELEASED` section under the appropriate heading with a link to the PR.
9. Celebrate :tada:!
