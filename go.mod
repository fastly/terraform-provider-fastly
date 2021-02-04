module github.com/fastly/terraform-provider-fastly

go 1.14

replace github.com/fastly/go-fastly/v2 v2.1.0 => github.com/opencredo/go-fastly/v2 v2.0.0-20210201113119-65d643a45b18

require (
	github.com/ajg/form v0.0.0-20160822230020-523a5da1a92f // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/fastly/go-fastly/v3 v3.0.0
	github.com/google/go-cmp v0.5.2
	github.com/google/jsonapi v0.0.0-20180313013858-2dcc18f43696 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk v1.1.0
	github.com/stretchr/testify v1.6.1
)
