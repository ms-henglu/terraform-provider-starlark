terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "addition" {
  value = provider::starlark::eval("x + y", { x = 5, y = 3 })
}
# Output: 8

output "multiplication" {
  value = provider::starlark::eval("x * y", { x = 10, y = 2 })
}
# Output: 20

output "floating_point" {
  value = provider::starlark::eval("a / b", { a = 10.0, b = 4.0 })
}
# Output: 2.5
