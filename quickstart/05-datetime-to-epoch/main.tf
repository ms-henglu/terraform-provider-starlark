terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

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
