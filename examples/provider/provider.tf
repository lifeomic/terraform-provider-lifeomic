terraform {
  required_providers {
    phc = {
      source  = "lifeomic/phc"
      version = "~> 1.0"
    }
  }
}

provider "phc" {
  account_id = "my-account"
}

