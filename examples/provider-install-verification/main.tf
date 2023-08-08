terraform {
  required_providers {
    gitbook = {
      source = "registry.terraform.io/gitbook/gitbook"
    }
  }
}

provider "gitbook" {}

data "gitbook_example" "example" {}