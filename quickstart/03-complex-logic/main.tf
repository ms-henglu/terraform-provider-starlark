terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "conditional_logic" {
  value = provider::starlark::eval(
    <<-EOT
    def get_status(code):
        if code >= 200 and code < 300:
            return "OK"
        elif code >= 400 and code < 500:
            return "Client Error"
        elif code >= 500:
            return "Server Error"
        else:
            return "Unknown"

    result = get_status(code)
    EOT
    ,
    { code = 404 }
  )
}
# Output: "Client Error"

output "list_filtering" {
  value = provider::starlark::eval(
    <<-EOT
        numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
        evens = [x for x in numbers if x % 2 == 0]
        result = evens
        EOT
    ,
    {} # No inputs needed
  )
}
# Output: [2, 4, 6, 8, 10]
