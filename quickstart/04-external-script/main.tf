terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "stats" {
  value = provider::starlark::eval(
    file("${path.module}/script.star"),
    { input_data = [10, 20, 30, 40, 50] }
  )
}
# Output:
# {
#   "average" = 30
#   "count"   = 5
#   "total"   = 150
# }
