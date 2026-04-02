variable "service_name" {
  type        = string
  description = "Name of the Fastly service."
}

variable "service_version" {
  type        = number
  description = "Service version to attach resources to."
}

variable "domain_name" {
  type        = string
  description = "Domain name to attach to the service."
}

variable "backends" {
  type = list(object({
    name    = string
    address = string
    port    = number
  }))
  description = "List of backends to attach."
}
