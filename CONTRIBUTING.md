# Contributing

We're happy to receive feature requests and PRs. If your change is
nontrivial, please open an
[issue](https://github.com/fastly/terraform-provider-fastly/issues/new)
to discuss the idea and implementation strategy before submitting a
PR.

1. Fork the repository.

1. Create an `upstream` remote.
```bash
$ git remote add upstream git@github.com:fastly/terraform-provider-fastly.git
```

1. Create a feature branch.

1. Make changes.

1. Write tests.

1. Validate your change via the steps documented [in the
   README](./README.md#testing).

1. Review the [documentation guide](./DOCUMENTATION.md) to ensure that
   you have properly documented your changes.

1. Open a pull request against `upstream main`. Note: once you have
   marked your PR as `Ready for Review` you should avoid 'force
   pushing' to the branch unless a reviewer asks you to do so.

1. Add an entry in `CHANGELOG.md` in the `UNRELEASED` section under
   the appropriate heading with a link to the PR.

1. Celebrate :tada:!
