resource "fastly_routing_config_domain_link" "example" {
    domain_id         = fastly_domain.example.domain_id
    routing_config_id = fastly_routing_config.example.routing_config_id
}
