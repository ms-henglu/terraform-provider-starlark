terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "calculation" {
  value = provider::starlark::eval(
    <<-EOT
    result = a + b
    EOT
    ,
    {
      a = 10
      b = 20
    }
  )
}
# Output: 30

output "parsed_list" {
  value = provider::starlark::eval(
    <<-EOT
    def get_list():
      return [1, 2, 3]
    
    result = get_list()
    EOT
    ,
    {}
  )
}
# Output: [1, 2, 3]

output "epoch_val" {
  value = provider::starlark::eval(
    <<-EOT
    def is_leap(y):
      return (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)

    def date_to_epoch(iso):
      # Format: 2023-01-01T00:00:00Z
      y = int(iso[0:4])
      m = int(iso[5:7])
      d = int(iso[8:10])
      h = int(iso[11:13])
      min = int(iso[14:16])
      s = int(iso[17:19])
      
      days = 0
      for cy in range(1970, y):
        days += 366 if is_leap(cy) else 365
      
      md = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
      if is_leap(y):
        md[1] = 29
        
      for cm in range(0, m-1):
        days += md[cm]
        
      days += d - 1
      return days * 86400 + h * 3600 + min * 60 + s

    result = str(date_to_epoch(v))
    EOT
    ,
    { v = "2023-01-01T00:00:00Z" }
  )
}
# Output: 1672531200

output "date_val" {
  value = provider::starlark::eval(
    <<-EOT
    def is_leap(y):
      return (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)

    def epoch_to_date(ts):
      ts = int(ts)
      days = ts // 86400
      rem = ts % 86400
      h = rem // 3600
      rem = rem % 3600
      min = rem // 60
      s = rem % 60
      
      y = 1970
      # Use a sufficient range for leap year calculation
      for i in range(1970, 3000):
        d_in_y = 366 if is_leap(i) else 365
        if days < d_in_y:
          y = i
          break
        days -= d_in_y
        
      md = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
      if is_leap(y):
        md[1] = 29
        
      m = 1
      for d_in_m in md:
        if days < d_in_m:
          break
        days -= d_in_m
        m += 1
        
      d = days + 1
      
      # formatting helper
      def pad(n):
        return "0" + str(n) if n < 10 else str(n)
        
      return str(y) + "-" + pad(m) + "-" + pad(d) + "T" + pad(h) + ":" + pad(min) + ":" + pad(s) + "Z"

    result = epoch_to_date(v)
    EOT
    ,
    { v = 1672531200 }
  )
}
# Output: "2023-01-01T00:00:00Z"

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

locals {
  # Define the script once
  epoch_to_date_script = <<-EOT
    def is_leap(y):
      return (y % 4 == 0 and y % 100 != 0) or (y % 400 == 0)

    def epoch_to_date(ts):
      ts = int(ts)
      days = ts // 86400
      rem = ts % 86400
      h = rem // 3600
      rem = rem % 3600
      min = rem // 60
      s = rem % 60
      
      y = 1970
      for i in range(1970, 3000):
        d_in_y = 366 if is_leap(i) else 365
        if days < d_in_y:
          y = i
          break
        days -= d_in_y
        
      md = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
      if is_leap(y):
        md[1] = 29
        
      m = 1
      for d_in_m in md:
        if days < d_in_m:
          break
        days -= d_in_m
        m += 1
        
      d = days + 1
      
      def pad(n):
        return "0" + str(n) if n < 10 else str(n)
        
      return str(y) + "-" + pad(m) + "-" + pad(d) + "T" + pad(h) + ":" + pad(min) + ":" + pad(s) + "Z"

    result = epoch_to_date(v)
  EOT
}

output "date_val_1" {
  value = provider::starlark::eval(local.epoch_to_date_script, { v = 1672531200 })
}
# Output: "2023-01-01T00:00:00Z"

output "date_val_2" {
  value = provider::starlark::eval(local.epoch_to_date_script, { v = 1704067200 })
}
# Output: "2024-01-01T00:00:00Z"
