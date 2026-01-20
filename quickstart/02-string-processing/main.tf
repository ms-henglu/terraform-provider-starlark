terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

output "uppercase" {
  value = provider::starlark::eval("s.upper()", { s = "hello world" })
}
# Output: "HELLO WORLD"

output "split_join" {
  value = provider::starlark::eval(
    <<-EOT
    words = s.split(",")
    result = " - ".join(words)
    EOT
    ,
    { s = "apple,banana,cherry" }
  )
}
# Output: "apple - banana - cherry"

output "formatting" {
    value = provider::starlark::eval(
        "'Hello {}'.format(name)",
        { name = "Terraform" }
    )
}
# Output: "Hello Terraform"
