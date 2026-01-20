# Parse CIDR Example

This example demonstrates how to implement a Bicep-like `parseCidr` function using Starlark. It accepts a CIDR string and returns network details like netmask, broadcast address, and usable IP range.

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Apply the configuration:
   ```bash
   terraform apply
   ```

3. Check the output for the calculated network details.
