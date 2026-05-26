# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

resource "fastly_service_cdn" "all_0" {
  provider = fastly
  comment  = "Managed by Terraform"
  name     = "Prod-Service"
}

import {
  to       = fastly_service_cdn.all_0
  provider = fastly
  identity = {
    service_id = "123456Prod"
  }
}

resource "fastly_service_cdn" "all_1" {
  provider = fastly
  comment  = "Managed by Terraform"
  name     = "Staging-Service"
}

import {
  to       = fastly_service_cdn.all_1
  provider = fastly
  identity = {
    service_id = "123456Stag"
  }
}

resource "fastly_service_cdn" "all_2" {
  provider = fastly
  comment  = "Managed by Terraform"
  name     = "Development-Service"
}

import {
  to       = fastly_service_cdn.all_2
  provider = fastly
  identity = {
    service_id = "123456Dev"
  }
}


# __generated__ by Terraform
resource "fastly_service_domain" "all_0" {
  provider   = fastly
  comment    = null
  name       = "prod.example.com"
  service_id = "123456Prod"
  version    = 37
}

import {
  to       = fastly_service_domain.all_0
  provider = fastly
  identity = {
    name       = "prod.example.com"
    service_id = "123456Prod"
    version    = 37
  }
}

resource "fastly_service_domain" "all_1" {
  provider   = fastly
  comment    = null
  name       = "staging.example.com"
  service_id = "123456Stag"
  version    = 22
}

import {
  to       = fastly_service_domain.all_1
  provider = fastly
  identity = {
    name       = "staging.example.com"
    service_id = "123456Stag"
    version    = 22
  }
}

resource "fastly_service_domain" "all_2" {
  provider   = fastly
  comment    = null
  name       = "dev.example.com"
  service_id = "123456Dev"
  version    = 207
}

import {
  to       = fastly_service_domain.all_2
  provider = fastly
  identity = {
    name       = "dev.example.com"
    service_id = "123456Dev"
    version    = 207
  }
}



# __generated__ by Terraform
resource "fastly_service_backend" "all_0" {
  provider   = fastly
  address    = "127.0.0.1"
  comment    = null
  name       = "Host 1"
  port       = 443
  service_id = "123456Prod"
  version    = 37
}

import {
  to       = fastly_service_backend.all_0
  provider = fastly
  identity = {
    name       = "Host 1"
    service_id = "123456Prod"
    version    = 37
  }
}

resource "fastly_service_backend" "all_1" {
  provider   = fastly
  address    = "127.0.0.1"
  comment    = null
  name       = "Host 2"
  port       = 80
  service_id = "123456Stag"
  version    = 22
}

import {
  to       = fastly_service_backend.all_1
  provider = fastly
  identity = {
    name       = "Host 2"
    service_id = "123456Stag"
    version    = 22
  }
}

resource "fastly_service_backend" "all_2" {
  provider   = fastly
  address    = "127.0.0.1"
  comment    = null
  name       = "Host 3"
  port       = 443
  service_id = "123456Dev"
  version    = 207
}

import {
  to       = fastly_service_backend.all_2
  provider = fastly
  identity = {
    name       = "Host 3"
    service_id = "123456Dev"
    version    = 207
  }
}
