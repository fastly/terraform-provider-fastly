resource "fastly_routing_config" "example" {
    name = "example-routing-config"

    path {
        path = "/api/*"

        rule {
            action_type  = "service"
            action_value = fastly_service_vcl.example.id
        }
    }

    path {
        path = "/"

        rule {
            action_type  = "service"
            action_value = fastly_service_vcl.example.id

            condition {
                type     = "header"
                key      = "X-Beta"
                operator = "equals"
                value    = "true"
            }
        }

        rule {
            action_type  = "service"
            action_value = fastly_service_vcl.example.id
        }
    }
}
