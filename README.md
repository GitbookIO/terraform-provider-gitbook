# Terraform GitBook Provider

The [GitBook
Provider](https://registry.terraform.io/providers/gitbook/gitbook/latest/docs)
allows [Terraform](https://terraform.io) syncing Terraform resources to the
[GitBook](https://gitbook.com) platform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Installation

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Usage

To get started, read the docs on
[GitBook](https://docs.gitbook.com/product-tour/integrations/terraform), or
check out the `gitbook` provider at the [Terraform
Registry](https://registry.terraform.io/providers/gitbook/gitbook/latest/docs).

## Development

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (see
[Requirements](#requirements) above).

Clone the repository, change to the directory, then run:

```sh
go install
```

This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

To generate or update documentation, run `go generate`.

## License

Unlicensed

---

© 2023 GitBook, Inc.
