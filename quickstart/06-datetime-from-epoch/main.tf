terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

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
      # Starlark has no while loop, use a sufficient range
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
