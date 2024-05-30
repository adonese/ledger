resource "aws_dynamodb_table" "UserBalanceTable" {
  name           = "UserBalanceTable"
  billing_mode   = "PROVISIONED"
  read_capacity  = 7
  write_capacity = 7
  hash_key       = "TenantID"
  range_key      = "AccountID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "AccountID"
    type = "S"
  }

  global_secondary_index {
    name               = "UserIndex"
    hash_key           = "AccountID"
    range_key          = "TenantID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }
}



resource "aws_dynamodb_table" "NilUsersTable" {
name           = "NilUsers"
  billing_mode   = "PROVISIONED"
  read_capacity  = 7
  write_capacity = 7
  hash_key       = "TenantID"
  range_key      = "AccountID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "AccountID"
    type = "S"
  }

  attribute {
    name = "Email"
    type = "S"
  }

  // Global Secondary Index on Email
  global_secondary_index {
    name               = "EmailIndex"
    hash_key           = "Email"
    range_key          = "TenantID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

  // Global Secondary Index on Username
  global_secondary_index {
    name               = "UsernameIndex"
    hash_key           = "AccountID"
    range_key          = "TenantID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }
}



resource "aws_ses_domain_identity" "example" {
  domain = "pynil.com"
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
  profile = "249"
}



resource "aws_dynamodb_table" "ledger_table" {
name           = "LedgerTable"
  billing_mode   = "PROVISIONED"
  read_capacity  = 7
  write_capacity = 7
  hash_key       = "TenantID"
  range_key      = "TransactionID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "TransactionID"
    type = "S"
  }

  global_secondary_index {
    name               = "TransactionIndex"
    hash_key           = "TenantID"
    range_key          = "TransactionID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }
}



resource "aws_dynamodb_table" "transactions" {
  name             = "TransactionsTable"
  billing_mode     = "PROVISIONED"
  read_capacity    = 7
  write_capacity   = 7
  hash_key         = "TenantID"
  range_key        = "TransactionID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "TransactionID"
    type = "S"
  }

  attribute {
    name = "TransactionDate"
    type = "N"
  }

  attribute {
    name = "FromAccount"
    type = "S"
  }

  attribute {
    name = "ToAccount"
    type = "S"
  }

  global_secondary_index {
    name               = "FromAccountIndex"
    hash_key           = "TenantID"
    range_key          = "FromAccount"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

  global_secondary_index {
    name               = "ToAccountIndex"
    hash_key           = "TenantID"
    range_key          = "ToAccount"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

  global_secondary_index {
    name               = "TransactionDateIndex"
    hash_key           = "TenantID"
    range_key          = "TransactionDate"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
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


# Create an S3 bucket
# resource "aws_s3_bucket" "image_bucket" {
#   bucket = "nil-kyc"
#   acl    = "private"

#   versioning {
#     enabled = true
#   }

#   server_side_encryption_configuration {
#     rule {
#       apply_server_side_encryption_by_default {
#         sse_algorithm = "AES256"
#       }
#     }
#   }
# }

# # Create an IAM user for accessing the bucket
# resource "aws_iam_user" "image_user" {
#   name = "image-user"
# }

# # Create an IAM access key for the user
# resource "aws_iam_access_key" "image_user_key" {
#   user = aws_iam_user.image_user.name
# }

# # Create an IAM policy for the user
# resource "aws_iam_user_policy" "image_user_policy" {
#   name = "image-user-policy"
#   user = aws_iam_user.image_user.name

#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [
#       {
#         Action = [
#           "s3:PutObject",
#           "s3:GetObject",
#           "s3:ListBucket"
#         ]
#         Effect   = "Allow"
#         Resource = [
#           "${aws_s3_bucket.image_bucket.arn}",
#           "${aws_s3_bucket.image_bucket.arn}/*"
#         ]
#       }
#     ]
#   })
# }

# # Output the access key and secret key
# output "access_key" {
#   value = aws_iam_access_key.image_user_key.id
# }

# output "secret_key" {
#   value     = aws_iam_access_key.image_user_key.secret
#   sensitive = true
# }


# # Write the access key and secret key to a file
# resource "local_file" "credentials" {
#   content  = <<-EOT
#     ACCESS_KEY=${aws_iam_access_key.image_user_key.id}
#     SECRET_KEY=${aws_iam_access_key.image_user_key.secret}
#   EOT
#   filename = "credentials.txt"
# }
