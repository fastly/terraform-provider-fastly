// Package constants holds shared literal values, notably the default log
// formats used as Computed schema defaults by the logging resources.
//
// The default-format constants intentionally end with a trailing newline (the
// closing "}" sits on its own line, followed by a blank line before the closing
// backtick). The Fastly API stores multi-line JSON formats with a trailing
// newline and returns them that way, so the constant needs to match it exactly.
// The Terraform plugin framework requires a Computed value's planned value to
// equal what the API returns after apply, so without the trailing newline the
// first apply fails with "Provider produced inconsistent result after apply".
package constants

// LoggingNewRelicOTLPDefaultFormat is the default log format for New Relic OTLP logging.
const LoggingNewRelicOTLPDefaultFormat = `{
  "timestamp":"%{strftime(\{"%Y-%m-%dT%H:%M:%S%z"\}, time.start)}V",
  "client_ip":"%{req.http.Fastly-Client-IP}V",
  "geo_country":"%{client.geo.country_name}V",
  "geo_city":"%{client.geo.city}V",
  "host":"%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
  "url":"%{json.escape(req.url)}V",
  "request_method":"%{json.escape(req.method)}V",
  "request_protocol":"%{json.escape(req.proto)}V",
  "request_referer":"%{json.escape(req.http.referer)}V",
  "request_user_agent":"%{json.escape(req.http.User-Agent)}V",
  "response_state":"%{json.escape(fastly_info.state)}V",
  "response_status":%{resp.status}V,
  "response_reason":%{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
  "response_body_size":%{resp.body_bytes_written}V,
  "fastly_server":"%{json.escape(server.identity)}V",
  "fastly_is_edge":%{if(fastly.ff.visits_this_service == 0, "true", "false")}V
}
`

// LoggingS3DefaultFormat is the default log format for S3 logging.
const LoggingS3DefaultFormat = `{
  "timestamp":"%{strftime(\{"%Y-%m-%dT%H:%M:%S%z"\}, time.start)}V",
  "client_ip":"%{req.http.Fastly-Client-IP}V",
  "geo_country":"%{client.geo.country_name}V",
  "geo_city":"%{client.geo.city}V",
  "host":"%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
  "url":"%{json.escape(req.url)}V",
  "request_method":"%{json.escape(req.method)}V",
  "request_protocol":"%{json.escape(req.proto)}V",
  "request_referer":"%{json.escape(req.http.referer)}V",
  "request_user_agent":"%{json.escape(req.http.User-Agent)}V",
  "response_state":"%{json.escape(fastly_info.state)}V",
  "response_status":%{resp.status}V,
  "response_reason":%{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
  "response_body_size":%{resp.body_bytes_written}V,
  "fastly_server":"%{json.escape(server.identity)}V",
  "fastly_is_edge":%{if(fastly.ff.visits_this_service == 0, "true", "false")}V
}
`
