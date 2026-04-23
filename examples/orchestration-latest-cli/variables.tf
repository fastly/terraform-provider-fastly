variable "fastly_api_key" {
  description = "Fastly API token."
  type        = string
}

variable "service_1_name" {
  description = "Name of Fastly service 1."
  type        = string
}

variable "service_2_name" {
  description = "Name of Fastly service 2."
  type        = string
}

variable "shared_backend" {
  description = "Shared backend configuration."
  type = object({
    name    = string
    address = string
    port    = number
  })
}
