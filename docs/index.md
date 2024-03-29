---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "lifeomic Provider"
subcategory: ""
description: |-
  
---

# lifeomic Provider



## Example Usage

```terraform
terraform {
  required_providers {
    lifeomic = {
      source  = "lifeomic/lifeomic"
      version = "~> 1.0"
    }
  }
}

provider "lifeomic" {
  account_id = "my-account"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `account_id` (String) The unique ID of the PHC Account to use this provider with. If not set explicitly in the provider block, `$LIFEOMIC_ACCOUNT` will be used.
- `headers` (Map of String) Additional headers that will be passed with any requests made. You can also use the LIFEOMIC_HEADERS environment variable as stringified JSON. Environment variables take precedent over other values
- `host` (String) The PHC API host to communicate with.. If not set explicitly in the provider block, `$LIFEOMIC_HOST` will be used.
- `token` (String, Sensitive) The token to use for authenticating with the PHC API. If not set explicitly in the provider block, `$LIFEOMIC_TOKEN` will be used.
