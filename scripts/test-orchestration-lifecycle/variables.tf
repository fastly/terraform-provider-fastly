variable "fastly_api_token" {
  description = "Fastly API token"
  type        = string
  sensitive   = true
}

variable "service_1_name" {
  description = "Name for service 1"
  type        = string
}

variable "service_1_version" {
  description = "Version number for service 1"
  type        = number
}

variable "service_1_domain" {
  description = "Domain name for service 1"
  type        = string
}

variable "service_2_name" {
  description = "Name for service 2"
  type        = string
}

variable "service_2_version" {
  description = "Version number for service 2"
  type        = number
}

variable "service_2_domain" {
  description = "Domain name for service 2"
  type        = string
}
