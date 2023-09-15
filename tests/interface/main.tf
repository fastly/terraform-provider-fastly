terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = ">1.0.0"
    }
  }
}

resource "fastly_service_vcl" "interface-test-project" {
  activate           = true
  comment            = "Fastly Terraform Provider: Interface Test Suite"
  default_host       = "interface-test-project.fastly-terraform.com"
  default_ttl        = 3600
  force_destroy      = true # omitted `reuse` as it conflicts
  http3              = false
  name               = "interface-test-project"
  stale_if_error     = false
  stale_if_error_ttl = 43200
  version_comment    = "Fastly Terraform Provider: Version comment example"

  acl {
    name = "test_acl"
  }

  backend {
    address = "127.0.0.1"
    name    = "test_backend"
    port    = 80
  }

  cache_setting {
    action          = "restart"
    cache_condition = "test_cache_condition"
    name            = "cache_backend"
    stale_ttl       = 1600
    ttl             = 300
  }

  condition {
    name      = "test_cache_condition"
    priority  = 20
    statement = "req.url ~ \"^/cache/\""
    type      = "CACHE"
  }

  condition {
    name      = "test_req_condition"
    priority  = 5
    statement = "req.url ~ \"^/foo/bar$\""
    type      = "REQUEST"
  }

  condition {
    name      = "test_res_condition"
    priority  = 10
    statement = "resp.status == 404"
    type      = "RESPONSE"
  }

  dictionary {
    name = "test_dictionary"
  }

  director {
    backends = ["test_backend"]
    name     = "test_director"
  }

  domain {
    comment = "demo"
    name    = "interface-test-project.fastly-terraform.com"
  }

  dynamicsnippet {
    name     = "test_dynamicsnippet"
    priority = 110
    type     = "recv"
  }

  gzip {
    content_types = ["application/x-javascript", "text/javascript"]
    extensions    = ["css"]
    name          = "all"
  }

  header {
    cache_condition    = "test_cache_condition"
    request_condition  = "test_req_condition"
    response_condition = "test_res_condition"
    action             = "set"
    destination        = "http.server-name"
    name               = "test_header"
    source             = "server.identity"
    type               = "request"
  }

  healthcheck {
    check_interval    = 4500
    expected_response = 404
    headers           = ["Beep: Boop"]
    host              = "example.com"
    http_version      = "1.0"
    initial           = 1
    method            = "POST"
    name              = "test_healthcheck"
    path              = "/test.txt"
    threshold         = 4
    timeout           = 4000
    window            = 10
  }

  logging_bigquery {
    account_name = "testloggingbigqueryaccountname"
    dataset      = "test_logging_bigquery_dataset"
    email        = "test_logging_bigquery@example.com"
    name         = "test_logging_bigquery"
    project_id   = "example-gcp-project"
    secret_key   = "<SECRET_KEY>"
    table        = "test_logging_bigquery_table"
    template     = "test_logging_bigquery_template"
  }
}
