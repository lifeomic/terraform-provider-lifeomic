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

