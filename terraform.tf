
resource "aws_dynamodb_table" "UserBalanceTable" {
  name           = "UserBalanceTable"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "AccountID"

  attribute {
    name = "AccountID"
    type = "S"
  }
}


resource "aws_dynamodb_table" "NilUsersTable" {
  name           = "NilUsers"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "AccountID"

  attribute {
    name = "AccountID"
    type = "S"
  }
}



resource "aws_ses_domain_identity" "example" {
  domain = "nil.sd"
}




terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region  = "us-east-1"
  profile = "default"
}



resource "aws_dynamodb_table" "ledger_table" {
  name           = "LedgerTable"
  billing_mode   = "PROVISIONED"
  read_capacity  = 20
  write_capacity = 20
  hash_key       = "AccountID"
  range_key      = "TransactionID"  # Add this line to set TransactionID as the sort key

  attribute {
    name = "AccountID"
    type = "S"
  }

  attribute {
    name = "TransactionID"
    type = "S"
  }

  # Adjust your GSI as needed. Depending on your access patterns, you might not need it anymore
  global_secondary_index {
    name               = "TransactionIndex"
    hash_key           = "TransactionID"
    write_capacity     = 20
    read_capacity      = 20
    projection_type    = "ALL"
  }./
}



resource "aws_dynamodb_table" "transactions" {
  name             = "TransactionsTable"
  billing_mode     = "PROVISIONED"
  read_capacity    = 20
  write_capacity   = 20
  hash_key         = "AccountID"
  range_key        = "TransactionDate"

  attribute {
    name = "AccountID"
    type = "S"
  }

  attribute {
    name = "TransactionDate"
    type = "N"  // Assuming the date is stored as a Unix timestamp
  }

  attribute {
    name = "ToAccount"
    type = "S"
  }

  attribute {
    name = "FromAccount"
    type = "S"
  }

  global_secondary_index {
    name               = "ToAccountIndex"
    hash_key           = "ToAccount"
    projection_type    = "ALL"
    read_capacity      = 20
    write_capacity     = 20
  }

  global_secondary_index {
    name               = "FromAccountIndex"
    hash_key           = "FromAccount"
    projection_type    = "ALL"
    read_capacity      = 20
    write_capacity     = 20
  }
}



# variable "github_token" {}

#resource "aws_amplify_app" "app" {
#  name = "nilpay"
#  repository = "https://github.com/nilpay/dashboard"
#  oauth_token = var.github_token
#}


#resource "aws_amplify_branch" "branch" {
#  app_id  = aws_amplify_app.app.id
#  branch_name = "main"
#}

