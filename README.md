# Fastly Terraform Provider

- Website: https://www.terraform.io
- Documentation: https://registry.terraform.io/providers/fastly/fastly/latest/docs
- Mailing list: http://groups.google.com/group/terraform-tool
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x or higher
-	[Go](https://golang.org/doc/install) 1.17 (to build the provider plugin)

> NOTE: the last version of the Fastly provider to support Terraform 0.11.x and below was [v0.26.0](https://github.com/fastly/terraform-provider-fastly/releases/tag/v0.26.0)

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/fastly/terraform-provider-fastly`

```sh
$ mkdir -p $GOPATH/src/github.com/fastly; cd $GOPATH/src/github.com/fastly
$ git clone git@github.com:fastly/terraform-provider-fastly
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/fastly/terraform-provider-fastly
$ make build
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.17+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in a local `bin` directory.

```sh
$ make build
...
```

Alongside the newly built binary a file called `developer_overrides.tfrc` will be created.  The `make build` target will communicate
back details for setting the `TF_CLI_CONFIG_FILE` environment variable that will enable Terraform to use your locally built provider binary.
* HashiCorp - [Development Overrides for Provider developers](https://www.terraform.io/docs/cli/config/config-file.html#development-overrides-for-provider-developers).

### Debugging the provider

The previous method with `dev_overrides` should be sufficient for most development use including testing your local changes with actual Terraform code.
However, sometimes it can be helpful to run the provider in debug mode if you need to attach a debugger like [delve](https://github.com/go-delve/delve) to solve a particular issue.

The way Terraform normally works is that it starts the provider in a subprocess and connects to it using GRPC over a local socket.
(For more information on this, see [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin#architecture)).
For debugging, it is possible to bypass this and execute the provider in a separate process, then tell Terraform how to communicate with it.
The benefit of this is that the provider can be launched with a debugger attached in the same way that the debugger would attach to any normal executable.

There are a few ways to do this depending on which debugger is being used, but we will use [delve](https://github.com/go-delve/delve) here as the process should be pretty similar for other debuggers.
The two things that need to be configured are that the provider is compiled without optimisations, and the executable is run with the `--debug` flag.
Compiling without optimisations ensures that the debugger can access all of the symbols in the binary that it needs to, and the `--debug` flag tells the Terraform plugin SDK to expect to be run in its own process, and to display the instructions for connecting to it after it starts up.

With [delve](https://github.com/go-delve/delve) this can be done in a single command:

```sh
$ dlv debug . -- --debug
Type 'help' for list of commands.
(dlv) continue
{"@level":"debug","@message":"plugin address","@timestamp":"2021-03-26T12:10:13.320981Z","address":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851","network":"unix"}
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:

        TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
```

This can also be done in two separate steps. The `-gcflags` disables optimisations (`-N`) and inlining (`-l`).

```sh
$ go build -gcflags="all=-N -l" -o terraform-provider-fastly_debug
$ dlv exec terraform-provider-fastly_debug -- --debug
```

As the message instructs, go to another shell and export the `TF_REATTACH_PROVIDERS` environment variable.
Then use Terraform as usual, and it will automatically use the provider in the debugger.

```sh
$ export TF_REATTACH_PROVIDERS='{"fastly/fastly":{"Protocol":"grpc","Pid":54132,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/qm/swg2hf4h5t8sdht8yhds4dg6m0000gn/T/plugin865249851"}}}'
$ terraform plan
```
You will then be able to set breakpoints and trace the provider's execution using the debugger as you would expect.

The implementation for setting up debug mode presumes Terraform 0.13.x is being used. If you're using Terraform 0.12.x you'll need to manually modify the value assigned to `TF_REATTACH_PROVIDERS` so that the key `"fastly/fastly"` becomes `"registry.terraform.io/-/fastly"`. See HashiCorp's ["Support for Debuggable Provider Binaries"](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries) for more details.

## Testing

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run. You should expect that the full acceptance test suite will take hours to run.

```sh
$ make testacc
```

In order to run an individual acceptance test, the '-run' flag can be used together with a regular expression.
The following example uses a regular expression matching single test called 'TestAccFastlyServiceVCL_basic'.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL_basic'
```

The following example uses a regular expression to execute a grouping of basic acceptance tests.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceVCL.*_basic'
```

In order to run the tests with extra debugging context, prefix the `make` command with `TF_LOG` (see the [terraform documentation](https://www.terraform.io/docs/internals/debugging.html) for details).

```sh
$ TF_LOG=trace make testacc
```

By default, the tests run with a parallelism of 4.
This can be reduced if some tests are failing due to network-related issues, or increased if possible, to reduce the running time of the tests.
Prefix the `make` command with `TEST_PARALLELISM`, as in the following example, to configure this.

```sh
$ TEST_PARALLELISM=8 make testacc
```

Depending on the Fastly account used, some features may not be enabled (e.g. Platform TLS).
This may result in some tests failing, potentially with `403 Unauthorised` errors, when the full test suite is being run.
Check the [Fastly API documentation](https://developer.fastly.com/reference/api/) to confirm if the failing tests use features in Limited Availability or only available to certain customers.
If this is the case, either use the `TESTARGS` regular expressions described above, or temporarily add `t.SkipNow()` to the top of any tests that should be excluded.

## Building The Documentation

The documentation is built from components (go templates) stored in the `templates` folder.
Building the documentation copies the full markdown into the `docs` folder, ready for deployment to Hashicorp.

> NOTE: you'll need the [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) tool for generating the Markdown to be deployed to Hashicorp. For more information on generating documentation, refer to https://www.terraform.io/docs/registry/providers/docs.html

* To validate the `/template` directory structure:
```
make validate-docs
```

* To build the `/docs` documentation Markdown files:
```
make generate-docs
```

* To view the documentation:
Paste `/docs` Markdown file content into https://registry.terraform.io/tools/doc-preview

## Contributing

Refer to [CONTRIBUTING.md](./CONTRIBUTING.md)
