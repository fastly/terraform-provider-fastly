resource "fastly_service_backend" "origin" {
  service_id            = fastly_service_cdn.test.id
  version               = {{.SERVICE_VERSION}}
  name                  = "{{.BACKEND_NAME}}"
  address               = "api.example.com"
  port                  = 443
  use_ssl               = true
  ssl_check_cert        = false
  ssl_cert_hostname     = "cert.example.com"
  ssl_sni_hostname      = "sni.example.com"
  ssl_client_secrets = {
    ssl_client_cert = <<-EOT
-----BEGIN CERTIFICATE-----
MIIDQTCCAimgAwIBAgIUKp4VT3Ue7gL5L3MIYbexM0+MQNIwDQYJKoZIhvcNAQEL
BQAwMDEuMCwGA1UEAwwldGVycmFmb3JtLXByb3ZpZGVyLWZhc3RseS10ZXN0LWNs
aWVudDAeFw0yNjA2MjkwOTU1NDJaFw0zNjA2MjYwOTU1NDJaMDAxLjAsBgNVBAMM
JXRlcnJhZm9ybS1wcm92aWRlci1mYXN0bHktdGVzdC1jbGllbnQwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDEZOCTK+jxA9SEBgpieJmX2F+vW8N+WaLw
QdHCV9JTa/2OkmfuX97Y00fDDcI5rArqbBAn9khOoVFvz+pRHOE/f/JyEGuoi2bp
hoC32fbAvbzmuNkX7Ho7/aNIGn1Baa986OfscYnfOAeehQk7SA/xJUXcMHsghlR/
uD53vMm9oNc7DjxU6QiKfrv0eAu5nmuPEkY+DL9kPJNe1B+9N6pXO2oRTSI5EACh
wc8yuvN6spJ/FeUxW4XFCeu2jQDk8Lqak48aqU+4z0YpH8B/DhtpQjK35ueLWFg8
iI9m2ybZrYjBqPV0MOBi0QSvvlp8+FmRbS3TtHZL8fvhqMHa5WF9AgMBAAGjUzBR
MB0GA1UdDgQWBBS3H2gAhvZnud8eq6OCwsgcEyBv9TAfBgNVHSMEGDAWgBS3H2gA
hvZnud8eq6OCwsgcEyBv9TAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUA
A4IBAQAmmF8dncBeaFKb3EakiPDIYJ7LNsJYhcrjx0pVBDB2dZ8Q9jz2sc5ezrT7
tayUnfkZL+hoW2YG2xrFaaQfJyHuGe30DLfkwI0+sYNSwTuITTA4eEc/nlhmsHF+
EcyhrT6Jmzacr2fq55NtSwweF1EJ8g1yCfIRmytUQlzXBdH6B3KnSDXs53uy/2eO
N2vwnZfyOYTxgVf2oN9v33KBvFM2esVVdHHDLb81MS0cr0z9Tz+JO04GzE8LNM48
2tFmuRwZKqTKZidYPpwIe4cuq9CboPBqeCqoI5Lb92GHyIMRUqH7OyMnGFKIPpuy
dJmR5YwzXAbyvuekuRVzwqcroKii
-----END CERTIFICATE-----
EOT
    ssl_client_key = <<-EOT
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDEZOCTK+jxA9SE
BgpieJmX2F+vW8N+WaLwQdHCV9JTa/2OkmfuX97Y00fDDcI5rArqbBAn9khOoVFv
z+pRHOE/f/JyEGuoi2bphoC32fbAvbzmuNkX7Ho7/aNIGn1Baa986OfscYnfOAee
hQk7SA/xJUXcMHsghlR/uD53vMm9oNc7DjxU6QiKfrv0eAu5nmuPEkY+DL9kPJNe
1B+9N6pXO2oRTSI5EAChwc8yuvN6spJ/FeUxW4XFCeu2jQDk8Lqak48aqU+4z0Yp
H8B/DhtpQjK35ueLWFg8iI9m2ybZrYjBqPV0MOBi0QSvvlp8+FmRbS3TtHZL8fvh
qMHa5WF9AgMBAAECggEADrKQS6CgtJrJW8kE8W9zlSrMUbgKa76tuZEUxNh5hSSN
i6QsHX7faKF6hcqA43vbLvB/5AcOdxNOYCcH8uKNqTN8KNT5mESNzYJd8y8KGaIJ
d0h+91ygnYk4NURFj/CRbBsVB5D2qmB2fNfuw9jl19w8Jm6aual7546vCXB04O6S
YfXWKt2YiqPYbGtn/50Us7J8X8tOHC107dcudFPen2RxB3HM+Nswe32nf+mkawob
nsJvOqvkLY48YeOe1PASofJVbQgnPTcMjuGhjjUPz9BOXo15uoq0vMFrK51+QR+Y
KkmqF7F+uq4m5pl9raVdaNP4aGkdD2n3L9P9dicaiQKBgQD4qFPd/bQjbifwJ6r4
4/ZiAVNeNgR9xrWzsJVIUgiOwzjkj2BYgIay610uS6yrXfgau5FX6s/1xtdgBe+6
rKb+BNkuIQKBNCQPjSqB0WP8Pg1q5paU3uWFR6Y+mx/y06nBHpyns08935oKDDV9
2fa0eV9Qha7OVeTYHhTLcsH91QKBgQDKMXmrTmb1z1EL/vzfPxacfRBKFIPhBwQX
RpQLfPdJYBANl83SmVqTIwuNVHOL6yHOxvSc8j4ByH3/klueFInodp719WFjocQg
hVaouAK5PZ3VPmYLJ8hWTAWHszHmSM66rTsYVXYuIL6O+KVeP/ZNao5rd405UA7J
jcfbPW8hCQKBgGURljUu/98+0QDuPrI3hlfDji1G64BsGkLVTXg9z7inZSKRnGmc
pCNpQ1Cj9aUZ5tSG1MbVbH3LupMPFqfbsWyib9wuEqSNmvKvQE3P3EIUvsNqwl30
U3pe6xWbW9sJaYBTfv0zBsxxbF0VJVDoHTyx8Kn8DFdV1lR5tZ4UIQGZAoGABRWZ
aaVfEW9VKmgPE84SU30Rm8tIRbBXef5cWq2Zyk6QGMdodZNFo82NzNAC19Hh18FJ
BWlSBdl00ahshV0e2qmg9a5l9Its0ySHOVbnOqFCBsq65izp7MGcofzvlErgZ/FT
OxlrD13jbNTz05roJqo3SpyHAJnyxT67d9fjo4ECgYAT5mLoXHFapK+ZsPCplnn/
kOso5hOtcQd7Aj6RZfppv62o2JgB6LowM8I46ope/BAz6c8/OOIsw0L+cH0rotIK
PO5xGr9Mzkaa7afJ85bbBJsfQlgV3Cb7h/He/iHNXBlLOgT1oN0sYIDNtFbJL94t
aiTOQNaNtvTztKWQwbjI0Q==
-----END PRIVATE KEY-----
EOT
  }
  min_tls_version       = "1.2"
  max_tls_version       = "1.3"
  ssl_ciphers           = "ECDHE-RSA-AES128-GCM-SHA256"
  weight                = 50
  max_conn              = 100
  connect_timeout       = 2000
  first_byte_timeout    = 10000
  between_bytes_timeout = 5000
  error_threshold       = 5
  max_lifetime          = 30000
  max_use               = 10
  override_host         = "override.example.com"
  shield                = "iad-va-us"
  auto_loadbalance      = true
  comment               = "Full test backend"
}
