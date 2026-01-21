# Terraform Provider Starlark

The `starlark` provider offers a set of functions that allow you to execute [Starlark](https://github.com/bazelbuild/starlark) (a dialect of Python) scripts within your Terraform configuration. This enables you to perform complex data transformations and logic that are not natively supported by Terraform functions.

## Features

* **Starlark Execution**: Inspect and control data flow with Python-like syntax (including loops and recursion) using the `eval` function.
* **Deterministic**: Operations are deterministic and side-effect free, ideal for Infrastructure-as-Code.
* **Zero Dependencies**: Simply include the script in your configuration or load it from a file.

## Example Usage

### Basic Calculation

```terraform
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
    "result = a + b",
    { a = 10, b = 20 }
  )
}
# Output: 30
```

### Advanced Usage

```terraform
output "logic" {
  value = provider::starlark::eval(
    <<-EOT
    def compute(val):
      if val > 100:
        return "high"
      return "low"
    
    result = compute(val)
    EOT
    ,
    { val = 150 }
  )
}
# Output: "high"
```

## Functions

*   [eval](docs/functions/eval.md): Executes the provided Starlark script with the given inputs.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.
