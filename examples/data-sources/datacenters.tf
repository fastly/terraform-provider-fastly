
data "fastly_datacenters" "fastly" {}

output "fastly_datacenters_all" {
  value = data.fastly_datacenters.fastly
}

output "fastly_datacenters_filtered" {
  # get the shield code of "TYO" POP
  value = one([for pop in data.fastly_datacenters.fastly.pops : pop.shield if pop["code"] == "TYO"])
}
