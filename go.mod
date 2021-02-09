module github.com/fastly/terraform-provider-fastly

go 1.14

replace github.com/fastly/go-fastly/v3 v3.0.0 => github.com/opencredo/go-fastly/v3 v3.0.0-20210209140016-2732f6b4c6d5

require (
	github.com/ajg/form v0.0.0-20160822230020-523a5da1a92f // indirect
	github.com/fastly/go-fastly/v3 v3.0.0
	github.com/google/go-cmp v0.5.2
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk v1.1.0
	github.com/stretchr/testify v1.6.1
)
