logging_s3 {
  name        = "{{.LOGGING_S3_NAME}}"
  bucket_name = "{{.BUCKET_NAME}}"
  authentication = {
    access_key = "AKIAIOSFODNN7EXAMPLE"
    secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  domain            = "s3.eu-west-1.amazonaws.com"
  path              = "/updated-logs/"
  period            = 1800
  gzip_level        = 9
  format_version    = 2
  message_type      = "classic"
  timestamp_format  = "%Y-%m-%dT%H:%M:%S%z"
  acl               = "public-read"
  redundancy        = "reduced_redundancy"
  processing_region = "eu"
  file_max_bytes    = 2097152
}
