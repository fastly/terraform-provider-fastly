logging_s3 {
  name        = "{{.LOGGING_S3_NAME}}"
  bucket_name = "{{.BUCKET_NAME}}"
  authentication = {
    access_key = "AKIAIOSFODNN7EXAMPLE"
    secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  domain            = "s3.us-west-2.amazonaws.com"
  path              = "/logs/"
  period            = 7200
  gzip_level        = 5
  format            = "%h %l %u %t \"%r\" %>s %b"
  format_version    = 1
  message_type      = "classic"
  timestamp_format  = "%Y-%m-%dT%H:%M:%S%z"
  acl               = "private"
  redundancy        = "standard"
  processing_region = "us"
  file_max_bytes    = 1048576
}
