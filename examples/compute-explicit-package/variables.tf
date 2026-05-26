variable "fastly_api_token" {
  description = "Fastly API token."
  type        = string
}

variable "service_name" {
  description = "Name of the Fastly Compute service."
  type        = string
}

variable "service_version" {
  description = "Writable service version to manage. For a new service this is usually 1."
  type        = number
}

variable "domain_name" {
  description = "Domain name to attach to the Compute service."
  type        = string
}

variable "backend" {
  description = "Backend configuration."
  type = object({
    name    = string
    address = string
    port    = number
  })
}

variable "package_filename" {
  description = "Path to the Compute package relative to this example directory."
  type        = string
}
