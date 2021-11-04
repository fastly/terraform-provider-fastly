# Terraform 0.13+ requires providers to be declared in a "required_providers" block
terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
      version >= "0.38.0"
    }
  }
}

# Configure the Fastly Provider
provider "fastly" {
  api_key = "test"
}

# Create a Service
resource "fastly_service_v1" "myservice" {
  name = "myawesometestservice"

  # ...
}