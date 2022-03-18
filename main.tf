# Example of how to use the resource

terraform {
  required_providers {
    phc = {
      version = "~> 1.0.0"
      source  = "lifeomic.com/tf/phc" # Doesn't mean anything
    }
  }
}

provider "phc" {
  account = "lifeomic"
  rules = {
    publishContent = true
  }
}

# Aliasing allows you have to clients to multiple accounts
provider "phc" {
  alias   = "lifeomiclife"
  account = "lifeomiclife"
  rules = {
    createData = true
    deleteData = true
    updateData = true
  }
}

locals {
  name        = "New Example Applet"
  description = "This is just used for verifying that the terraform provider works. This should never appear in prod"
  app_url     = "https://lifeapplets.dev.lifeomic.com/anxiety"
}

resource "phc_applet" "terraform_test_applet" {
  provider       = phc.lifeomiclife
  name           = local.name
  description    = local.description
  author_display = "LifeOmic"
  url            = local.app_url
  image          = "${local.app_url}/icon-240.png"
}

resource "phc_app_tile" "example_applet" {
  name         = local.name
  description  = local.description
  image        = "icon-240.png" # Git ignored, feel free to put whatever here when testing
  image_hash   = filemd5("./icon-240.png")
  app_tile_id  = phc_applet.terraform_test_applet.id
  auto_version = true
  lifecycle {
    ignore_changes = [
      image,
      version,
    ]
  }
}
