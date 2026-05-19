fastly_api_key = "<your_token>"

# Service 1 (new service)
service_1_name    = "service-1"
service_1_version = 2

# Service 2 (new service)
service_2_name    = "service-2"
service_2_version = 1

# Shared backend
shared_backend = {
  name    = "shared-origin"
  address = "shared.origin.example.com"
  port    = 443
}
