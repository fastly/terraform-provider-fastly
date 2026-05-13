terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {
  api_key = var.fastly_api_key
}

resource "fastly_service_vcl_explicit" "service_1" {
  name    = var.service_1_name
  comment = "Managed by Terraform demo provider"
}

resource "fastly_service_vcl_explicit" "service_2" {
  name    = var.service_2_name
  comment = "Managed by Terraform demo provider"
}

data "fastly_service_version" "service_1" {
  service_id = fastly_service_vcl_explicit.service_1.id
}

data "fastly_service_version" "service_2" {
  service_id = fastly_service_vcl_explicit.service_2.id
}

module "service_1" {
  source          = "./modules/service_config"
  service_id      = fastly_service_vcl_explicit.service_1.id
  service_version = data.fastly_service_version.service_1.latest_version
  domain_name     = "www.service1.example.com"

  backends = concat(
    [var.shared_backend],
    [
      {
        name    = "unique-origin-1"
        address = "unique1.origin.example.foo.com"
        port    = 80
      }
    ]
  )
}

module "service_2" {
  source          = "./modules/service_config"
  service_id      = fastly_service_vcl_explicit.service_2.id
  service_version = data.fastly_service_version.service_2.latest_version
  domain_name     = "www.service2.example.com"

  backends = [var.shared_backend]
}
