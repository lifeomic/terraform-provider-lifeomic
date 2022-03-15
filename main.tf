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
  rules = {
    publishContent = true
  }
}

resource "app_tile" "example_applet" {
  provider     = phc
  name         = "Example Applet"
  description  = "This is just used for verifying that the terraform provider works. This should never appear in prod"
  image        = "icon-240.png" # Git ignored, feel free to put whatever here when testing
  image_hash   = filemd5("./icon-240.png")
  app_tile_id  = "58e9ede8-eb28-40b6-82a6-d8b670d9c651" # An example dev id that happens to exist
  auto_version = true
  lifecycle {
    ignore_changes = [
      image,
      version,
    ]
  }
}
