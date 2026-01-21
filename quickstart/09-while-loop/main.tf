terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "fibonacci_sequence" {
  value = provider::starlark::eval(
    <<-EOT
    def fibonacci(limit):
        a, b = 0, 1
        result = []
        while a < limit:
            result.append(a)
            a, b = b, a + b
        return result

    result = fibonacci(limit)
    EOT
    ,
    { limit = 100 }
  )
}
# Output: [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89]
