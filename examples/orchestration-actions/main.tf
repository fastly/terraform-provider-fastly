terraform {
  required_providers {
    fastly = {
      source  = "example/fastly"
      version = "0.1.0"
    }
  }
}

provider "fastly" {
  api_key = var.fastly_api_key
}

# Service 1 (shared + unique backend)
module "service_1" {
  source          = "./modules/service"
  service_name    = var.service_1_name
  service_version = var.service_1_version
  domain_name     = "www.service1.example.com"

  backends = concat(
    [var.shared_backend],
    [
      {
        name    = "unique-origin-1"
        address = "unique1.origin.example.foo.com"
        port    = 443
      }
    ]
  )
}

# Service 2 (shared backend only)
module "service_2" {
  source          = "./modules/service"
  service_name    = var.service_2_name
  service_version = var.service_2_version
  domain_name     = "www.service2.example.com"

  backends = [var.shared_backend]
}


data "fastly_service_version" "service_1" {
  service_id = module.service_1.service_id
}

data "fastly_service_version" "service_2" {
  service_id = module.service_2.service_id
}

# -------------------------------
# Actions (version lifecycle)
# -------------------------------

# Clone current versions (invoked explicitly)
action "fastly_service_version_clone" "service_1" {
  config {
    service_id = module.service_1.service_id
    version    = data.fastly_service_version.service_1.active_version
  }
}

action "fastly_service_version_clone" "service_2" {
  config {
    service_id = module.service_2.service_id
    version    = data.fastly_service_version.service_2.active_version
  }
}

# Activate versions (invoked explicitly)
action "fastly_service_version_activate" "service_1_prod" {
  config {
    service_id = module.service_1.service_id
    version    = var.service_1_version
    staging    = false
  }
}

action "fastly_service_version_activate" "service_1_staging" {
  config {
    service_id = module.service_1.service_id
    version    = var.service_1_version
    staging    = true
  }
}

action "fastly_service_version_activate" "service_2_prod" {
  config {
    service_id = module.service_2.service_id
    version    = var.service_2_version
    staging    = false
  }
}

action "fastly_service_version_activate" "service_2_staging" {
  config {
    service_id = module.service_2.service_id
    version    = var.service_2_version
    staging    = true
  }
}
