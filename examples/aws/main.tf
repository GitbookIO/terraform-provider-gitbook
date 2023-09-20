terraform {
  required_providers {
    gitbook = {
      source = "registry.terraform.io/gitbook/gitbook"
    }
  }
}

provider "gitbook" {}

locals {
  organization_id    = "4Me7JapjYF3sgxrFoKxP"
  account_arn        = "arn:aws:organizations::123456789012:account/o-123456/123456789012"
  my_function_arn    = "arn:aws:lambda:us-east-2:123456789012:function:my-function"
  my_function_region = "us-east-2"
}

resource "gitbook_entity_schema" "aws_account" {
  organization_id = local.organization_id
  title = {
    singular : "AWS Account"
    plural : "AWS Accounts",
  }
  type = "terraform:aws-account"
  properties = [
    {
      "name" : "arn",
      "title" : "ARN",
      "type" : "text"
    },
    {
      "name" : "id",
      "title" : "ID",
      "type" : "text"
    },
    {
      "name" : "email",
      "title" : "Email",
      "type" : "text"
    }
  ]
}

resource "gitbook_entity_schema" "aws_lambda_function" {
  organization_id = local.organization_id
  type            = "terraform:aws-lambda-function"
  title = {
    singular = "AWS Lambda Function"
    plural   = "AWS Lambda Functions",
  }
  properties = [
    {
      "name" : "arn",
      "title" : "ARN",
      "type" : "text"
    },
    {
      "entity" : {
        "type" : gitbook_entity_schema.aws_account.type
      },
      "name" : "account",
      "title" : "Account",
      "type" : "relation"
    },
    {
      "name" : "region",
      "title" : "Region",
      "type" : "text"
    }
  ]
}

resource "gitbook_entity" "example_aws_lambda_function" {
  organization_id = gitbook_entity_schema.aws_lambda_function.organization_id
  type            = gitbook_entity_schema.aws_lambda_function.type
  entity_id       = local.my_function_arn
  properties = {
    account = {
      relation = {
        entity_id = local.account_arn
      }
    },
    arn = {
      string = local.my_function_arn
    },
    region = {
      string = local.my_function_region
    }
  }
}
