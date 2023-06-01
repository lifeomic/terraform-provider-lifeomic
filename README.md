# terraform-provider-lifeomic

The Terraform LifeOmic provider is a plugin for Terraform that allows for the full
lifecycle management of LifeOmic resources.

## Usage

See the [LifeOmic Provider documentation](https://registry.terraform.io/providers/lifeomic/lifeomic)
to get started using the LifeOmic provider. There are also some basic [examples](./examples/)
in this repository to demonstrate usage.

## Development

### Building the provider

In order to build the provider from source, you'll need to have [go][go-binaries]1.19+
installed. Then run `make build`.

### Regenerating GQL Client

If you're picking up changes to GQL APIs, run `make generate`.

If it's not downloading schema updates try running `make -B generate`

### Using a local provider build

Refer to the upstream documentation on [development overrides][tf-dev-overrides].

### Running acceptance tests

In order to run acceptance test, you must first [obtain an auth token][auth-token-guide].
Set the `LIFEOMIC_TOKEN` environment variable to your token and `LIFEOMIC_ACCOUNT` to your
account's unique identifier.

```shell
LIFEOMIC_TOKEN=<auth-token> LIFEOMIC_ACCOUNT=<account-id> make acctest
```

[go-binaries]: https://go.dev/dl/
[tf-dev-overrides]: https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers
[auth-token-guide]: https://platform.docs.lifeomic.com/user-guides/account-info/api-keys

#### Wellness Offering Acceptance Tests

(must be logged in to AWS)

```shell
 LIFEOMIC_TOKEN=fake-token LIFEOMIC_ACCOUNT=fake-account TESTARGS="-run='TestAccMarketplaceWellnessOffering'" LIFEOMIC_USE_LAMBDA=1 make acctest
```
