# Create a ConfigStore to hold entries
resource "fastly_configstore" "my_store" {
  name = "my-config-store"
}

# Add a single entry to the ConfigStore
resource "fastly_configstore_entry" "example_entry" {
  store_id = fastly_configstore.my_store.id
  key      = "api_endpoint"
  value    = "https://api.example.com/v1"
}

# Add another entry with a different key
resource "fastly_configstore_entry" "another_entry" {
  store_id = fastly_configstore.my_store.id
  key      = "timeout_seconds"
  value    = "30"
}