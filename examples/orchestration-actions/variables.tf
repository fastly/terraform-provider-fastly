variable "fastly_api_key" {
  description = "Fastly API token."
  type        = string
}

# Service 1
variable "service_1_name" {
  description = "Name of Fastly service 1."
  type        = string
}

variable "service_1_version" {
  description = "Version number for Service 1 (must exist in Fastly)."
  type        = number
}

# Service 2
variable "service_2_name" {
  description = "Name of Fastly service 2."
  type        = string
}

variable "service_2_version" {
  description = "Version number for Service 2 (must exist in Fastly)."
  type        = number
}

# Shared backend definition
variable "shared_backend" {
  description = "Shared backend configuration."
  type = object({
    name    = string
    address = string
    port    = number
  })
}
