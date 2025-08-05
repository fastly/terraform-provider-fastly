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
  force_destroy      = true # Omitted `reuse` as it conflicts
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
    name      = "test_prefetch_condition"
    priority  = 15
    statement = "req.url~+\"index.html\""
    type      = "PREFETCH"
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
    format       = <<-EOT
{
    "timestamp": "%%{strftime({"%Y-%m-%dT%H:%M:%S"}, time.start)}V",
    "client_ip": "%%{req.http.Fastly-Client-IP}V",
    "geo_country": "%%{client.geo.country_name}V",
    "geo_city": "%%{client.geo.city}V",
    "host": "%%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
    "url": "%%{json.escape(req.url)}V",
    "request_method": "%%{json.escape(req.method)}V",
    "request_protocol": "%%{json.escape(req.proto)}V",
    "request_referer": "%%{json.escape(req.http.referer)}V",
    "request_user_agent": "%%{json.escape(req.http.User-Agent)}V",
    "response_state": "%%{json.escape(fastly_info.state)}V",
    "response_status": %%{resp.status}V,
    "response_reason": %%{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
    "response_body_size": %%{resp.body_bytes_written}V,
    "fastly_server": "%%{json.escape(server.identity)}V",
    "fastly_is_edge": %%{if(fastly.ff.visits_this_service == 0, "true", "false")}V
  }
EOT
  }

  product_enablement {
    bot_management     = false
    brotli_compression = true
    domain_inspector   = false
    image_optimizer    = false
    origin_inspector   = false
    websockets         = false
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    feature_revision     = 1
    http_methods         = "POST,PUT,PATCH,DELETE"
    logger_type          = "bigquery"
    name                 = "test_rate_limiter"
    penalty_box_duration = 30

    response {
      content      = "test_rate_limiter_content"
      content_type = "plain/text"
      status       = 429
    }

    response_object_name = "test_rate_limiter_response_object"
    rps_limit            = 10
    # uri_dictionary_name  = "test_dictionary" # Omitted as dictionary needs to exist before this is executed
    window_size = 60
  }

  request_setting {
    action            = "pass"
    bypass_busy_wait  = true
    default_host      = "interface-test-project.fastly-terraform.com"
    force_miss        = true
    force_ssl         = false
    hash_keys         = "req.url.path, req.http.host" # Omitted because of error... Syntax error: Expected string variable or constant
    max_stale_age     = "300"
    name              = "test_request_setting"
    request_condition = "test_req_condition"
    timer_support     = true
    xff               = "append"
  }

  response_object {
    cache_condition   = "test_cache_condition"
    content           = "test content"
    content_type      = "text/html"
    name              = "test_response_object"
    request_condition = "test_req_condition"
    response          = "OK"
    status            = 200
  }

  snippet {
    content  = "if ( req.url ) { set req.http.different-header = \"true\"; }"
    name     = "recv_test"
    priority = 110
    type     = "recv"
  }

  vcl {
    content = "# some vcl here"
    main    = true
    name    = "test_vcl"
  }
}
