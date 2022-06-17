module github.com/fastly/terraform-provider-fastly

go 1.16

require (
	github.com/bflad/tfproviderlint v0.27.1
	github.com/fastly/go-fastly/v6 v6.3.2
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-docs v0.5.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f
)

replace github.com/fastly/go-fastly/v6 => github.com/noseglid/go-fastly/v6 v6.0.0-20220616150832-4a299a27ec6b
