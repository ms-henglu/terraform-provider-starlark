terraform {
  required_providers {
    starlark = {
      source = "ms-henglu/starlark"
    }
  }
}

provider "starlark" {}

locals {
  base_config = {
    settings = {
      enabled = true
      size    = "medium"
      color   = "blue"
    }
    tags = {
      env = "dev"
    }
  }

  override_config = {
    settings = {
      color = "red"
    }
    tags = {
      owner = "team-a"
    }
  }
}

output "merged" {
  value = provider::starlark::eval(
    <<-EOT
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
    EOT
    ,
    {
      map1 = local.base_config
      map2 = local.override_config
    }
  )
}
