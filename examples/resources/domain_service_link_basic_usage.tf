import {
    to = fastly_domain_service_link.example
    id = "%s"
}

resource "fastly_domain_service_link" "example" {
    domain_id = "%s"
    service_id = "%s"
}