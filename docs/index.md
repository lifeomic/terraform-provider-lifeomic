# <provider> Provider

This provider is for managing LifeOmic PHC resources.

The provider uses your local AWS config in order to authenticate. Support for token-based authentication with the public graphql-proxy will probably come in the future.

## Example Usage

```hcl
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
  image        = "icon-240.png"
  image_hash   = filemd5("./icon-240.png")
  app_tile_id  = phc_applet.terraform_test_applet.id
  auto_version = true
  lifecycle {
    ignore_changes = [
      image,   # Hash is what actually matters
      version, # Autoversioned
    ]
  }
}
```

## Argument Reference

### Provider Args
* account: string # Determines what account you are using for these resources
* rules: string # Principle of least privilege, you should specify only the privileges you need
* user: string # Allows specifying a specific user. Defaults to phc-tf. Useful for log searching

### phc_applet
* name: string
* description: string
* author_display: string
* image: string # URL to hosted image
* url: string

### phc_app_tile

* name: string
* description: string
* app_tile_id: string
* image: string # Path to image
* image_hash: string # Hash so that we know when the image has changed
* version: string
* auto_version: bool # Will autoincrement the patch value on any change

