terraform {
  required_providers {
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly" {
  api_token = var.fastly_api_token
}

# Service 1
resource "fastly_service_cdn" "service_1" {
  name          = var.service_1_name
  comment       = "Test service 1"
  force_destroy = true
}

resource "fastly_service_domain" "service_1_domain" {
  service_id = fastly_service_cdn.service_1.id
  version    = var.service_1_version
  name       = var.service_1_domain
}

resource "fastly_service_backend" "service_1_backend_shared" {
  service_id        = fastly_service_cdn.service_1.id
  version           = var.service_1_version
  name              = "shared-origin"
  address           = "shared.origin.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "shared.origin.example.com"
  ssl_sni_hostname  = "shared.origin.example.com"
}

resource "fastly_service_backend" "service_1_backend_unique" {
  service_id        = fastly_service_cdn.service_1.id
  version           = var.service_1_version
  name              = "unique-origin-1"
  address           = "unique1.origin.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "unique1.origin.example.com"
  ssl_sni_hostname  = "unique1.origin.example.com"
}

resource "fastly_service_cdn_acl" "service_1_acl" {
  service_id    = fastly_service_cdn.service_1.id
  version       = var.service_1_version
  name          = "test_acl_1"
  force_destroy = true
}

resource "fastly_service_cdn_acl_entries" "service_1_acl_entries" {
  service_id     = fastly_service_cdn.service_1.id
  acl_id         = fastly_service_acl.service_1_acl.acl_id
  manage_entries = true

  entry {
    ip      = "192.168.1.0"
    subnet  = "24"
    negated = false
    comment = "Service 1 test entry"
  }

  entry {
    ip      = "10.0.0.0"
    subnet  = "8"
    negated = true
    comment = "Service 1 blocked network"
  }
}

# Optional domain and backend for testing version writes
resource "fastly_service_domain" "service_1_new_domain" {
  count      = var.service_1_new_domain != "" ? 1 : 0
  service_id = fastly_service_cdn.service_1.id
  version    = var.service_1_version
  name       = var.service_1_new_domain
}

resource "fastly_service_backend" "service_1_new_backend" {
  count             = var.service_1_new_backend != "" ? 1 : 0
  service_id        = fastly_service_cdn.service_1.id
  version           = var.service_1_version
  name              = "new-origin"
  address           = var.service_1_new_backend
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = var.service_1_new_backend
  ssl_sni_hostname  = var.service_1_new_backend
}

# Service 2
resource "fastly_service_cdn" "service_2" {
  name          = var.service_2_name
  comment       = "Test service 2"
  force_destroy = true
}

resource "fastly_service_domain" "service_2_domain" {
  service_id = fastly_service_cdn.service_2.id
  version    = var.service_2_version
  name       = var.service_2_domain
}

resource "fastly_service_backend" "service_2_backend_shared" {
  service_id        = fastly_service_cdn.service_2.id
  version           = var.service_2_version
  name              = "shared-origin"
  address           = "shared.origin.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "shared.origin.example.com"
  ssl_sni_hostname  = "shared.origin.example.com"
}

resource "fastly_service_cdn_acl" "service_2_acl" {
  service_id    = fastly_service_cdn.service_2.id
  version       = var.service_2_version
  name          = "test_acl_2"
  force_destroy = true
}

resource "fastly_service_cdn_acl_entries" "service_2_acl_entries" {
  service_id     = fastly_service_cdn.service_2.id
  acl_id         = fastly_service_acl.service_2_acl.acl_id
  manage_entries = true

  entry {
    ip      = "172.16.0.0"
    subnet  = "12"
    negated = false
    comment = "Service 2 test entry"
  }
}

# Data sources to check version state
data "fastly_service_version" "service_1" {
  service_id = fastly_service_cdn.service_1.id
  depends_on = [
    fastly_service_domain.service_1_domain,
    fastly_service_backend.service_1_backend_shared,
    fastly_service_backend.service_1_backend_unique,
    fastly_service_cdn_acl.service_1_acl,
    fastly_service_cdn_acl_entries.service_1_acl_entries
  ]
}

data "fastly_service_version" "service_2" {
  service_id = fastly_service_cdn.service_2.id
  depends_on = [
    fastly_service_domain.service_2_domain,
    fastly_service_backend.service_2_backend_shared,
    fastly_service_cdn_acl.service_2_acl,
    fastly_service_cdn_acl_entries.service_2_acl_entries
  ]
}

# Actions for version lifecycle management
action "fastly_service_version_clone" "service_1_clone" {
  config {
    service_id = fastly_service_cdn.service_1.id
    version    = data.fastly_service_version.service_1.active_version
  }
}

action "fastly_service_version_clone" "service_1_clone_from_latest" {
  config {
    service_id = fastly_service_cdn.service_1.id
    version    = data.fastly_service_version.service_1.latest_version
  }
}

action "fastly_service_version_clone" "service_2_clone" {
  config {
    service_id = fastly_service_cdn.service_2.id
    version    = data.fastly_service_version.service_2.active_version
  }
}

# Clones whichever version resources currently point at (var.service_N_version),
# used to move off a locked version before destroy without relying on the
# active/latest version happening to match.
action "fastly_service_version_clone" "service_1_clone_from_pinned" {
  config {
    service_id = fastly_service_cdn.service_1.id
    version    = var.service_1_version
  }
}

action "fastly_service_version_clone" "service_2_clone_from_pinned" {
  config {
    service_id = fastly_service_cdn.service_2.id
    version    = var.service_2_version
  }
}

action "fastly_service_version_activate" "service_1_activate" {
  config {
    service_id = fastly_service_cdn.service_1.id
    version    = var.service_1_version
  }
}

action "fastly_service_version_activate" "service_2_activate" {
  config {
    service_id = fastly_service_cdn.service_2.id
    version    = var.service_2_version
  }
}
