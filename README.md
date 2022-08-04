# terraform-provider-phc

The Terraform PHC provider is a plugin for Terraform that allows for the full
lifecycle managment of PHC resources.

## Usage

See the [PHC Provider documentation](https://registry.terraform.io/providers/lifeomic/phc)
to get started using the PHC prvoider. There are also some basic [examples](./examples/)
in this repository to demonstrate usage.

## Development

### Building the provider

In order to build the provider from source, you'll need to have [go][go-binaries]1.19+
installed. Then run `make build`.

### Using a local provider build

Refer to the upstream documentation on [development overrides][tf-dev-overrides].

### Running acceptance tests

In order to run acceptance test, you must first [obtain an auth token][auth-token-guide].
Set the `PHC_TOKEN` environment variable to your token and `PHC_ACCOUNT` to your
account's unique identifier.

```shell
PHC_TOKEN=<auth-token> PHC_ACCOUNT=<account-id> make acctest
```

[go-binaries]: https://go.dev/dl/
[tf-dev-overrides]: https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers
[auth-token-guide]: https://phc.docs.lifeomic.com/user-guides/account-management/api-keys

