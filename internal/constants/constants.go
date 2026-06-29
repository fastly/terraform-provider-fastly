package constants

// LoggingS3DefaultFormat is the default log format for S3 logging.
const LoggingS3DefaultFormat = `{
    "timestamp": "%{strftime(\{"%Y-%m-%dT%H:%M:%S%z"\}, time.start)}V",
    "client_ip": "%{req.http.Fastly-Client-IP}V",
    "geo_country": "%{client.geo.country_name}V",
    "geo_city": "%{client.geo.city}V",
    "host": "%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
    "url": "%{json.escape(req.url)}V",
    "request_method": "%{json.escape(req.method)}V",
    "request_protocol": "%{json.escape(req.proto)}V",
    "request_referer": "%{json.escape(req.http.referer)}V",
    "request_user_agent": "%{json.escape(req.http.User-Agent)}V",
    "response_state": "%{json.escape(fastly_info.state)}V",
    "response_status": %{resp.status}V,
    "response_reason": %{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
    "response_body_size": %{resp.body_bytes_written}V,
    "fastly_server": "%{json.escape(server.identity)}V",
    "fastly_is_edge": %{if(fastly.ff.visits_this_service == 0, "true", "false")}V
  }`
