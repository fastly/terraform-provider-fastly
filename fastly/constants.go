package fastly

// MessageTypeDescription describes the message format.
const MessageTypeDescription = "How the message should be formatted. Can be either `classic`, `loggly`, `logplex` or `blank`. Default is `classic`"

// GzipLevelDescription describes Gzip compression.
const GzipLevelDescription = "Level of Gzip compression from `0-9`. `0` means no compression. `1` is the fastest and the least compressed version, `9` is the slowest and the most compressed version. Default `0`"

// TimestampFormatDescription describes the timestamp format.
const TimestampFormatDescription = "The `strftime` specified timestamp formatting (default `%Y-%m-%dT%H:%M:%S.000`)"

// SnippetTypeDescription describes the VCL snippet location.
const SnippetTypeDescription = "The location in generated VCL where the snippet should be placed (can be one of `init`, `recv`, `hash`, `hit`, `miss`, `pass`, `fetch`, `error`, `deliver`, `log` or `none`)"

// LoggingFormatUpdate is the generic logging format used for tests.
const LoggingFormatUpdate = "%h %l %u %t \"%r\" %>s %b"

// LoggingBigQueryDefaultFormat - is the default format for BigQuery logging.
const LoggingBigQueryDefaultFormat = `{
    "timestamp": "%{strftime(\{"%Y-%m-%dT%H:%M:%S"\}, time.start)}V",
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

// LoggingBlobStorageDefaultFormat - is the default format for Blob Storage logging.
const LoggingBlobStorageDefaultFormat = `{
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

// LoggingCloudFilesDefaultFormat - is the default format for CloudFiles logging.
const LoggingCloudFilesDefaultFormat = `{
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

// LoggingDatadogDefaultFormat - is the default format for Datadog logging.
const LoggingDatadogDefaultFormat = `{
    "ddsource": "fastly",
    "service": "%{req.service_id}V",
    "date": "%{begin:%Y-%m-%dT%H:%M:%S%z}t",
    "time_start": "%{begin:%Y-%m-%dT%H:%M:%S%Z}t",
    "time_end": "%{end:%Y-%m-%dT%H:%M:%S%Z}t",
    "http": {
      "request_time_ms": %{time.elapsed.msec}V,
      "method": "%m",
      "url": "%{json.escape(req.url)}V",
      "useragent": "%{json.escape(req.http.User-Agent)}V",
      "referer": "%{json.escape(req.http.referer)}V",
      "protocol": "%H",
      "request_x_forwarded_for": "%{X-Forwarded-For}i",
      "status_code": "%s"
    },
    "network": {
      "client": {
       "ip": "%h",
       "name": "%{client.as.name}V",
       "number": "%{client.as.number}V",
       "connection_speed": "%{client.geo.conn_speed}V"
      },
     "destination": {
       "ip": "%A"
      },
    "geoip": {
    "geo_city": "%{client.geo.city.utf8}V",
    "geo_country_code": "%{client.geo.country_code}V",
    "geo_continent_code": "%{client.geo.continent_code}V",
    "geo_region": "%{client.geo.region}V"
    },
    "bytes_written": %B,
    "bytes_read": %{req.body_bytes_read}V
    },
    "host":"%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
    "origin_host": "%v",
    "is_ipv6": %{if(req.is_ipv6, "true", "false")}V,
    "is_tls": %{if(req.is_ssl, "true", "false")}V,
    "tls_client_protocol": "%{json.escape(tls.client.protocol)}V",
    "tls_client_servername": "%{json.escape(tls.client.servername)}V",
    "tls_client_cipher": "%{json.escape(tls.client.cipher)}V",
    "tls_client_cipher_sha": "%{json.escape(tls.client.ciphers_sha)}V",
    "tls_client_tlsexts_sha": "%{json.escape(tls.client.tlsexts_sha)}V",
    "is_h2": %{if(fastly_info.is_h2, "true", "false")}V,
    "is_h2_push": %{if(fastly_info.h2.is_push, "true", "false")}V,
    "h2_stream_id": "%{fastly_info.h2.stream_id}V",
    "request_accept_content": "%{Accept}i",
    "request_accept_language": "%{Accept-Language}i",
    "request_accept_encoding": "%{Accept-Encoding}i",
    "request_accept_charset": "%{Accept-Charset}i",
    "request_connection": "%{Connection}i",
    "request_dnt": "%{DNT}i",
    "request_forwarded": "%{Forwarded}i",
    "request_via": "%{Via}i",
    "request_cache_control": "%{Cache-Control}i",
    "request_x_requested_with": "%{X-Requested-With}i",
    "request_x_att_device_id": "%{X-ATT-Device-Id}i",
    "content_type": "%{Content-Type}o",
    "is_cacheable": %{if(fastly_info.state~"^(HIT|MISS)$", "true","false")}V,
    "response_age": "%{Age}o",
    "response_cache_control": "%{Cache-Control}o",
    "response_expires": "%{Expires}o",
    "response_last_modified": "%{Last-Modified}o",
    "response_tsv": "%{TSV}o",
    "server_datacenter": "%{server.datacenter}V",
    "req_header_size": %{req.header_bytes_read}V,
    "resp_header_size": %{resp.header_bytes_written}V,
    "socket_cwnd": %{client.socket.cwnd}V,
    "socket_nexthop": "%{client.socket.nexthop}V",
    "socket_tcpi_rcv_mss": %{client.socket.tcpi_rcv_mss}V,
    "socket_tcpi_snd_mss": %{client.socket.tcpi_snd_mss}V,
    "socket_tcpi_rtt": %{client.socket.tcpi_rtt}V,
    "socket_tcpi_rttvar": %{client.socket.tcpi_rttvar}V,
    "socket_tcpi_rcv_rtt": %{client.socket.tcpi_rcv_rtt}V,
    "socket_tcpi_rcv_space": %{client.socket.tcpi_rcv_space}V,
    "socket_tcpi_last_data_sent": %{client.socket.tcpi_last_data_sent}V,
    "socket_tcpi_total_retrans": %{client.socket.tcpi_total_retrans}V,
    "socket_tcpi_delta_retrans": %{client.socket.tcpi_delta_retrans}V,
    "socket_ploss": %{client.socket.ploss}V
  }`

// LoggingDigitalOceanDefaultFormat - is the default format for DigitalOcean logging.
const LoggingDigitalOceanDefaultFormat = `%h %l %u %t "%r" %>s %b`

// LoggingElasticsearchDefaultFormat - is the default format for Elasticsearch logging.
const LoggingElasticsearchDefaultFormat = `%h %l %u %t "%r" %>s %b`

// LoggingFTPDefaultFormat - is the default format for FTP logging.
const LoggingFTPDefaultFormat = `{
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
}`

// LoggingGCSDefaultFormat - is the default format for Google Cloud Storage logging.
const LoggingGCSDefaultFormat = `{
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

// LoggingGooglePubSubDefaultFormat - is the default format for Google Pub/Sub logging.
const LoggingGooglePubSubDefaultFormat = `{
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

// LoggingGrafanaCloudLogsDefaultFormat - is the default format for Grafana Cloud Logs logging.
const LoggingGrafanaCloudLogsDefaultFormat = `{
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

// LoggingHerokuDefaultFormat - is the default format for Heroku logging.
const LoggingHerokuDefaultFormat = `{
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

// LoggingHoneycombDefaultFormat - is the default format for Honeycomb logging.
const LoggingHoneycombDefaultFormat = `{
    "time":"%{begin:%Y-%m-%dT%H:%M:%SZ}t",
    "data":  {
      "service_id":"%{req.service_id}V",
      "time_elapsed":%D,
      "request":"%m",
      "host":"%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
      "url":"%{cstr_escape(req.url)}V",
      "protocol":"%H",
      "is_ipv6":%{if(req.is_ipv6, "true", "false")}V,
      "is_tls":%{if(req.is_ssl, "true", "false")}V,
      "is_h2":%{if(fastly_info.is_h2, "true", "false")}V,
      "client_ip":"%h",
      "geo_city":"%{client.geo.city.utf8}V",
      "geo_country_code":"%{client.geo.country_code}V",
      "server_datacenter":"%{server.datacenter}V",
      "request_referer":"%{Referer}i",
      "request_user_agent":"%{User-Agent}i",
      "request_accept_content":"%{Accept}i",
      "request_accept_language":"%{Accept-Language}i",
      "request_accept_charset":"%{Accept-Charset}i",
      "cache_status":"%{regsub(fastly_info.state, "^(HIT-(SYNTH)|(HITPASS|HIT|MISS|PASS|ERROR|PIPE)).*", "\\2\\3") }V",
      "status":"%s",
      "content_type":"%{Content-Type}o",
      "req_header_size":%{req.header_bytes_read}V,
      "req_body_size":%{req.body_bytes_read}V,
      "resp_header_size":%{resp.header_bytes_written}V,
      "resp_body_size":%{resp.body_bytes_written}V
    }
  }`

// LoggingHTTPSDefaultFormat - is the default format for HTTPS logging.
const LoggingHTTPSDefaultFormat = `{
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

// LoggingKafkaDefaultFormat - is the default format for Kafka logging.
const LoggingKafkaDefaultFormat = `{
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

// LoggingKinesisDefaultFormat - is the default format for Kinesis logging.
const LoggingKinesisDefaultFormat = `{
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

// LoggingLogentriesDefaultFormat - is the default format for Logentries logging
// ** This seems to be deprecated **
// const LoggingLogentriesDefaultFormat = ``

// LoggingLogglyDefaultFormat - is the default format for Loggly logging.
const LoggingLogglyDefaultFormat = `{
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

// LoggingLogshuttleDefaultFormat - is the default format for Logshuttle logging.
const LoggingLogshuttleDefaultFormat = `{
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

// LoggingNewRelicDefaultFormat - is the default format for New Relic logging.
const LoggingNewRelicDefaultFormat = `{
    "timestamp": %{time.start.msec}V,
    "logtype": "accesslogs",
    "cache_status":"%{regsub(fastly_info.state, "^(HIT-(SYNTH)|(HITPASS|HIT|MISS|PASS|ERROR|PIPE)).*", "\2\3") }V",
    "client_ip":"%h",
    "client_device_type":"%{client.platform.hwtype}V",
    "client_os_name":"%{client.os.name}V",
    "client_os_version":"%{client.os.version}V",
    "client_browser_name":"%{client.browser.name}V",
    "client_browser_version":"%{client.browser.version}V",
    "client_as_name":"%{client.as.name}V",
    "client_as_number":"%{client.as.number}V",
    "client_connection_speed": "%{client.geo.conn_speed}V",
    "client_port": %{client.port}V,
    "client_rate_bps":%{client.socket.tcpi_delivery_rate}V,
    "client_recv_bytes":%{client.socket.tcpi_bytes_received}V,
    "client_requests_count":%{client.requests}V,
    "client_resp_body_size_write": %{resp.body_bytes_written}V,
    "client_resp_header_size_write": %{resp.header_bytes_written}V,
    "client_resp_ttfb": %{time.to_first_byte}V,
    "client_rtt_us":%{client.socket.tcpi_rtt}V,
    "content_type":"%{Content-Type}o",
    "domain": "%{Fastly-Orig-Host}i",
    "fastly_datacenter": "%{server.datacenter}V",
    "fastly_host": "%{server.hostname}V",
    "fastly_is_edge": %{if(fastly.ff.visits_this_service == 0, "true", "false")}V,
    "fastly_region": "%{server.region}V",
    "fastly_server": "%{json.escape(server.identity)}V",
    "host": "%v",
    "origin_host":"%v",
    "origin_name":"%{req.backend.name}V",
    "request":"%{req.request}V",
    "request_method":"%m",
    "request_accept_charset":"%{json.escape(req.http.Accept-Charset)}V",
    "request_accept_language":"%{json.escape(req.http.Accept-Language)}V",
    "request_referer":"%{json.escape(req.http.Referer)}V",
    "request_user_agent":"%{json.escape(req.http.User-Agent)}V",
    "resp_status":"%s",
    "response": "%{resp.response}V",
    "service_id":"%{req.service_id}V",
    "service_version": "%{req.vcl.version}V",
    "status":"%s",
    "time_start":"%{begin:%Y-%m-%dT%H:%M:%SZ}t",
    "time_end":"%{end:%Y-%m-%dT%H:%M:%SZ}t",
    "time_elapsed":%D,
    "tls_cipher":"%{json.escape(tls.client.cipher)}V",
    "tls_version":"%{json.escape(tls.client.protocol)}V",
    "url":"%{json.escape(req.url)}V",
    "user_agent":"%{json.escape(req.http.User-Agent)}V",
    "user_city":"%{client.geo.city.utf8}V",
    "user_country_code":"%{client.geo.country_code}V",
    "user_continent_code":"%{client.geo.continent_code}V",
    "user_region":"%{client.geo.region}V"
  }`

// LoggingNewRelicOLTPDefaultFormat - is the default format for New Relic logging.
const LoggingNewRelicOLTPDefaultFormat = `{
     "timestamp":"%{strftime({"%Y-%m-%dT%H:%M:%S%z"}, time.start)}V",
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
   }`

// LoggingOpenStackDefaultFormat - is the default format for OpenStack logging.
const LoggingOpenStackDefaultFormat = `{
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

// LoggingPapertrailDefaultFormat - is the default format for Papertrail logging.
const LoggingPapertrailDefaultFormat = `{
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

// LoggingS3DefaultFormat - is the default format for S3 logging.
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

// LoggingScalyrDefaultFormat - is the default format for Scalyr logging.
const LoggingScalyrDefaultFormat = `{
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

// LoggingSFTPDefaultFormat - is the default format for SFTP logging.
const LoggingSFTPDefaultFormat = `{
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

// LoggingSplunkDefaultFormat - is the default format for Splunk logging.
const LoggingSplunkDefaultFormat = `{
    "time":%{time.start.sec}V,
    "host":"%{Fastly-Orig-Host}i",
    "event": {
      "service_id":"%{req.service_id}V",
      "time_start":"%{begin:%Y-%m-%dT%H:%M:%S%Z}t",
      "time_end":"%{end:%Y-%m-%dT%H:%M:%S%Z}t",
      "time_elapsed":%D,
      "client_ip":"%h",
      "client_as_name":"%{client.as.name}V",
      "client_as_number":"%{client.as.number}V",
      "client_connection_speed":"%{client.geo.conn_speed}V",
      "request":"%m",
      "protocol":"%H",
      "origin_host":"%v",
      "url":"%{json.escape(req.url)}V",
      "is_ipv6":%{if(req.is_ipv6, "true", "false")}V,
      "is_tls":%{if(req.is_ssl, "true", "false")}V,
      "tls_client_protocol":"%{json.escape(tls.client.protocol)}V",
      "tls_client_servername":"%{json.escape(tls.client.servername)}V",
      "tls_client_cipher":"%{json.escape(tls.client.cipher)}V",
      "tls_client_cipher_sha":"%{json.escape(tls.client.ciphers_sha )}V",
      "tls_client_tlsexts_sha":"%{json.escape(tls.client.tlsexts_sha)}V",
      "is_h2":%{if(fastly_info.is_h2, "true", "false")}V,
      "is_h2_push":%{if(fastly_info.h2.is_push, "true", "false")}V,
      "h2_stream_id":"%{fastly_info.h2.stream_id}V",
      "request_referer":"%{Referer}i",
      "request_user_agent":"%{User-Agent}i",
      "request_accept_content":"%{Accept}i",
      "request_accept_language":"%{Accept-Language}i",
      "request_accept_encoding":"%{Accept-Encoding}i",
      "request_accept_charset":"%{Accept-Charset}i",
      "request_connection":"%{Connection}i",
      "request_dnt":"%{DNT}i",
      "request_forwarded":"%{Forwarded}i",
      "request_via":"%{Via}i",
      "request_cache_control":"%{Cache-Control}i",
      "request_x_requested_with":"%{X-Requested-With}i",
      "request_x_att_device_id":"%{X-ATT-Device-Id}i",
      "request_x_forwarded_for":"%{X-Forwarded-For}i",
      "status":"%s",
      "content_type":"%{Content-Type}o",
      "cache_status":"%{regsub(fastly_info.state, "^(HIT-(SYNTH)|(HITPASS|HIT|MISS|PASS|ERROR|PIPE)).*", "\\2\\3")}V",
      "is_cacheable":%{if(fastly_info.state ~"^(HIT|MISS)$", "true", "false")}V,
      "response_age":"%{Age}o",
      "response_cache_control":"%{Cache-Control}o",
      "response_expires":"%{Expires}o",
      "response_last_modified":"%{Last-Modified}o",
      "response_tsv":"%{TSV}o",
      "server_datacenter":"%{server.datacenter}V",
      "server_ip":"%A",
      "geo_city":"%{client.geo.city.utf8}V",
      "geo_country_code":"%{client.geo.country_code}V",
      "geo_continent_code":"%{client.geo.continent_code}V",
      "geo_region":"%{client.geo.region}V",
      "req_header_size":%{req.header_bytes_read}V,
      "req_body_size":%{req.body_bytes_read}V,
      "resp_header_size":%{resp.header_bytes_written}V,
      "resp_body_size":%B,
      "socket_cwnd":%{client.socket.cwnd}V,
      "socket_nexthop":"%{client.socket.nexthop}V",
      "socket_tcpi_rcv_mss":%{client.socket.tcpi_rcv_mss}V,
      "socket_tcpi_snd_mss":%{client.socket.tcpi_snd_mss}V,
      "socket_tcpi_rtt":%{client.socket.tcpi_rtt}V,
      "socket_tcpi_rttvar":%{client.socket.tcpi_rttvar}V,
      "socket_tcpi_rcv_rtt":%{client.socket.tcpi_rcv_rtt}V,
      "socket_tcpi_rcv_space":%{client.socket.tcpi_rcv_space}V,
      "socket_tcpi_last_data_sent":%{client.socket.tcpi_last_data_sent}V,
      "socket_tcpi_total_retrans":%{client.socket.tcpi_total_retrans}V,
      "socket_tcpi_delta_retrans":%{client.socket.tcpi_delta_retrans}V,
      "socket_ploss":%{client.socket.ploss}V
    }
  }`

// LoggingSumologicDefaultFormat - is the default format for Sumologic logging.
const LoggingSumologicDefaultFormat = `{
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

// LoggingSyslogDefaultFormat - is the default format for Syslog logging.
const LoggingSyslogDefaultFormat = `{
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
