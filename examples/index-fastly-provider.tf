# Terraform 0.13+ requires providers to be declared in a "required_providers" block
terraform {
  required_providers {
    fastly = {
      source  = "fastly/fastly"
      version = ">= 01c6f4f714f1d026713b9f25ba47c7c139b16731"
    }
  }
}

# Configure the Fastly Provider
provider "fastly" {
  api_key = "test"
}

# Create a Service
resource "fastly_service_vcl" "myservice" {
  name = "myawesometestservice"

  # ...
}