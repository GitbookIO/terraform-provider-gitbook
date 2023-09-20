# Terraform GitBook Provider

The [GitBook
Provider](https://registry.terraform.io/providers/gitbook/gitbook/latest/docs)
allows [Terraform](https://terraform.io) syncing Terraform resources to the
[GitBook](https://gitbook.com) platform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Usage

To get started, read the docs on [GitBook](https://docs.gitbook.com/product-tour/integrations/terraform), or check out the `gitbook` provider at
the [Terraform Registry](https://registry.terraform.io/providers/gitbook/gitbook/latest/docs).

## Development

For local development on the provider, follow these steps:

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

To generate or update documentation, run `go generate`.

Follow [these docs](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
to configure using the locally installed provider.

## License

Unlicensed

---

© 2023 GitBook, Inc.
