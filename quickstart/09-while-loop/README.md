# While Loop Example

This example demonstrates how to use `while` loops in Starlark scripts, a feature enabled by this provider.

Starlark often restricts indefinite loops to ensure termination, but the Terraform Starlark Provider enables `while` loops to support more complex algorithms.

## Main Function

The script calculates the Fibonacci sequence up to a given limit using a `while` loop.

```python
def fibonacci(limit):
    a, b = 0, 1
    result = []
    while a < limit:
        result.append(a)
        a, b = b, a + b
    return result

result = fibonacci(limit)
```

## Usage

Run the example:

```bash
terraform init
terraform apply
```

## Output

The output `fibonacci_sequence` contains the list of Fibonacci numbers less than 100.
