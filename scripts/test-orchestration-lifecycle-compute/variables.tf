variable "fastly_api_token" {
  type        = string
  description = "Fastly API token for authentication"
  sensitive   = true
}

variable "service_1_name" {
  type        = string
  description = "Name for test compute service 1"
}

variable "service_1_domain" {
  type        = string
  description = "Domain name for test compute service 1"
}

variable "service_1_version" {
  type        = number
  description = "Version number for service 1 resources"
  default     = 1
}

variable "service_2_name" {
  type        = string
  description = "Name for test compute service 2"
}

variable "service_2_domain" {
  type        = string
  description = "Domain name for test compute service 2"
}

variable "service_2_version" {
  type        = number
  description = "Version number for service 2 resources"
  default     = 1
}

variable "package_path" {
  type        = string
  description = "Path to the compute package tar.gz file"
}
