resource "fastly_domain" "example" {
    fqdn = "example.com"
    service_id = "12345abcde"
    description = "This is a test domain."
}
