Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.14 (to build the provider plugin)

Building The Provider
---------------------

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

Using the provider
----------------------
## Fill in for each provider

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.14+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-fastly
...
```

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
The following example uses a regular expression matching single test called 'TestAccFastlyServiceV1_basic'.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceV1_basic'
```

The following example uses a regular expression to execute a grouping of basic acceptance tests.

```sh
$ make testacc TESTARGS='-run=TestAccFastlyServiceV1_.*_basic'
```

Building The Documentation
--------------------------

The documentation is built from components (go templates) stored in the `website_src` folder.
Building the documentation copies the full markdown into the `website` folder, ready for deployment to Hashicorp.

With the repository cloned to: `$GOPATH/src/github.com/fastly/terraform-provider-fastly`:

* To build the documentation:
`go run scripts/website/parse-templates.go `

* To build and preview the documentation online:
`make website`

Contributing
--------------------------

Refer to [CONTRIBUTING.md](./CONTRIBUTING.md)
