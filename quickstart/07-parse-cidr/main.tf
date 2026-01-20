terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "cidr_obj" {
  value = provider::starlark::eval(
    <<-EOT
    def ip_to_int(ip):
      parts = ip.split('.')
      return (int(parts[0]) << 24) + (int(parts[1]) << 16) + (int(parts[2]) << 8) + int(parts[3])

    def int_to_ip(val):
      return str((val >> 24) & 0xFF) + "." + str((val >> 16) & 0xFF) + "." + str((val >> 8) & 0xFF) + "." + str(val & 0xFF)

    def parse_cidr(cidr_str):
      parts = cidr_str.split('/')
      ip_int = ip_to_int(parts[0])
      prefix = int(parts[1])
      
      wildcard = (1 << (32 - prefix)) - 1
      mask = ~wildcard & 0xFFFFFFFF
      
      network_int = ip_int & mask
      broadcast_int = network_int | wildcard
      first_int = network_int + 1
      last_int = broadcast_int - 1
      
      return {
        "network": int_to_ip(network_int),
        "netmask": int_to_ip(mask),
        "broadcast": int_to_ip(broadcast_int),
        "firstUsable": int_to_ip(first_int),
        "lastUsable": int_to_ip(last_int),
      }

    result = parse_cidr(v)
    EOT
    ,
    { v = "10.0.0.0/24" }
  )
}
# Output:
# {
#   "broadcast" = "10.0.0.255"
#   "firstUsable" = "10.0.0.1"
#   "lastUsable" = "10.0.0.254"
#   "netmask" = "255.255.255.0"
#   "network" = "10.0.0.0"
# }
