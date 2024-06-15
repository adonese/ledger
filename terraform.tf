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

  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

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
    name = "UUID"
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

    global_secondary_index {
    name               = "UserUUIDIndex"
    hash_key           = "TenantID"
    range_key          = "UUID"
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
    name = "UUID"
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

    global_secondary_index {
    name               = "UserUUIDIndex"
    hash_key           = "TenantID"
    range_key          = "UUID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }
}


# This is for backing up our data. We don't want to inadvertently delete important data
resource "aws_dynamodb_table" "DeletedNilUsers" {
  name           = "DeletedNilUsers"
  billing_mode   = "PAY_PER_REQUEST"
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
}


resource "aws_iam_role" "lambda_dynamodb_role" {
  name = "lambda_dynamodb_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com",
        },
      },
    ],
  })
}

resource "aws_iam_policy" "lambda_dynamodb_policy" {
  name        = "lambda_dynamodb_policy"
  description = "Policy for Lambda to access DynamoDB"
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem",
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:DescribeStream",
          "dynamodb:ListStreams",
        ],
        Effect = "Allow",
        Resource = [
          aws_dynamodb_table.NilUsersTable.arn,
          aws_dynamodb_table.NilUsersTable.stream_arn,
          aws_dynamodb_table.DeletedNilUsers.arn,
        ],
      },
    ],
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attach" {
  role       = aws_iam_role.lambda_dynamodb_role.name
  policy_arn = aws_iam_policy.lambda_dynamodb_policy.arn
}

data "local_file" "lambda_zip_hash" {
  filename = "function.zip"
}

resource "aws_lambda_function" "dynamodb_stream_processor" {
  filename         = "function.zip"
  function_name    = "dynamodb_stream_processor"
  role             = aws_iam_role.lambda_dynamodb_role.arn
  handler          = "bootstrap"
  runtime          = "provided.al2"
  source_code_hash = filebase64sha256("function.zip")

  environment {
    variables = {
      DESTINATION_TABLE = aws_dynamodb_table.DeletedNilUsers.name
    }
  }
}

resource "aws_lambda_event_source_mapping" "dynamodb_stream_mapping" {
  event_source_arn  = aws_dynamodb_table.NilUsersTable.stream_arn
  function_name     = aws_lambda_function.dynamodb_stream_processor.arn
  starting_position = "LATEST"
}


# for qr based payments
resource "aws_dynamodb_table" "qr_payments_table" {
  name           = "QRPaymentsTable"
  billing_mode   = "PROVISIONED"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "TenantID"
  range_key      = "PaymentID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "PaymentID"
    type = "S"
  }

  attribute {
    name = "UUID"
    type = "S"
  }

  attribute {
    name = "Status"
    type = "S"
  }


  attribute {
    name = "AccountID"
    type = "S"
  }

  global_secondary_index {
    name               = "UUIDIndex"
    hash_key           = "TenantID"
    range_key          = "UUID"
    projection_type    = "ALL"
    read_capacity      = 5
    write_capacity     = 5
  }

  global_secondary_index {
    name               = "StatusIndex"
    hash_key           = "TenantID"
    range_key          = "Status"
    projection_type    = "ALL"
    read_capacity      = 5
    write_capacity     = 5
  }

  global_secondary_index {
    name               = "AccountIDIndex"
    hash_key           = "TenantID"
    range_key          = "AccountID"
    projection_type    = "ALL"
    read_capacity      = 5
    write_capacity     = 5
  }
}



# Escrow data 
resource "aws_dynamodb_table" "escrow_meta" {
name           = "EscrowMeta"
  billing_mode   = "PROVISIONED"
  read_capacity  = 7
  write_capacity = 7
  hash_key       = "TenantID"

  attribute {
    name = "TenantID"
    type = "S"
  }

  attribute {
    name = "Webhook"
    type = "S"
  }

    global_secondary_index {
    name               = "WebhookIndex"
    hash_key           = "TenantID"
    range_key          = "Webhook"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }
}


resource "aws_dynamodb_table" "escrow_transactions" {
  name             = "EscrowTransactions"
  billing_mode     = "PROVISIONED"
  read_capacity    = 7
  write_capacity   = 7
  hash_key         = "UUID"
  range_key        = "TransactionID"

  attribute {
    name = "UUID"
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
    hash_key           = "UUID"
    range_key          = "FromAccount"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

  global_secondary_index {
    name               = "ToAccountIndex"
    hash_key           = "UUID"
    range_key          = "ToAccount"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }


  global_secondary_index {
    name               = "TransactionDateIndex"
    hash_key           = "UUID"
    range_key          = "TransactionDate"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

    global_secondary_index {
    name               = "SystemID"
    hash_key           = "TransactionID"
    range_key          = "UUID"
    projection_type    = "ALL"
    read_capacity      = 7
    write_capacity     = 7
  }

  stream_enabled = true
  stream_view_type = "NEW_AND_OLD_IMAGES"


}


resource "aws_iam_role" "lambda_exec_role" {
  name = "lambda_exec_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "lambda.amazonaws.com",
        },
      },
    ],
  })
}

resource "aws_iam_role_policy" "lambda_exec_policy" {
  name = "lambda_exec_policy"
  role = aws_iam_role.lambda_exec_role.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "dynamodb:Scan",
          "dynamodb:Query",
          "dynamodb:GetItem",
          "dynamodb:BatchGetItem",
          "dynamodb:DescribeStream",
          "dynamodb:GetRecords",
          "dynamodb:GetShardIterator",
          "dynamodb:ListStreams"
        ],
        Effect = "Allow",
        Resource = [
          "${aws_dynamodb_table.escrow_transactions.arn}",
          "${aws_dynamodb_table.escrow_transactions.arn}/stream/*"
        ],
      },
      {
        Action = "logs:*",
        Effect = "Allow",
        Resource = "arn:aws:logs:*:*:*",
      },
    ],
  })
}

resource "aws_lambda_function" "escrow_transaction_processor" {
  filename         = "bootstrap"
  function_name    = "escrow_transaction_processor"
  role             = aws_iam_role.lambda_exec_role.arn
  handler          = "main"
  runtime          = "provided.al2023"
  source_code_hash = filebase64sha256("bootstrap.zip")
}

resource "aws_lambda_event_source_mapping" "dynamodb_stream" {
  event_source_arn  = aws_dynamodb_table.escrow_transactions.stream_arn
  function_name     = aws_lambda_function.escrow_transaction_processor.arn
  starting_position = "LATEST"
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
