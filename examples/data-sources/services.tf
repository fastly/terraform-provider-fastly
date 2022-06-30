data "fastly_services" "services" {}

output "fastly_services_all" {
  value = data.fastly_services.services
}

output "fastly_services_filtered" {
  # get the service with the name "Example Service"
  value = one([for service in data.fastly_services.services.details : service.id if service.name == "Example Service"])
}
