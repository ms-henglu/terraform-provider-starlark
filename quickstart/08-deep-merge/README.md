# Deep Merge Example

This example demonstrates how to implement a **deep merge** function using Starlark.

Terraform's built-in `merge()` function is shallow, meaning it replaces top-level keys completely. For managing complex configurations (like overriding specific fields in a nested structure), a deep merge is often required.

## Main Function

The Starlark script defines a `deep_merge` function that recursively merges two dictionaries.

```python
def deep_merge(a, b):
    res = {}
    for k, v in a.items():
        res[k] = v
        
    for k, v in b.items():
        if k in res and type(res[k]) == "dict" and type(v) == "dict":
            res[k] = deep_merge(res[k], v)
        else:
            res[k] = v
    return res

result = deep_merge(map1, map2)
```

## Usage

Run the example:

```bash
terraform init
terraform apply
```

## Output

The output `merged` will show the combined configuration where `base_config` values are preserved unless specifically overridden by `override_config`.
