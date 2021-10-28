# Configure the Fastly Provider
provider "fastly" {
api_key = "test"
}

# Create a Service
resource "fastly_service_v1" "myservice" {
name = "myawesometestservice"

# ...
}