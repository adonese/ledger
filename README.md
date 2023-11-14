# API Documentation for Ledger Package

## Overview

The `ledger` package is designed to manage user balances and transactions within a ledger system. It uses AWS DynamoDB for storage and integrates with AWS SES for email notifications.

## Use the package

You can import the package directly, `go get github.com/adonese/ledger@latest`

## Authentication

To use the API, authenticate with AWS by initializing a DynamoDB client using `InitializeLedger`.

### InitializeLedger

```go
func InitializeLedger(accessKey, secretKey, region string) (*dynamodb.Client, error)
```

**Parameters:**
- `accessKey`: AWS access key.
- `secretKey`: AWS secret key.
- `region`: AWS region.

**Returns:**
- `*dynamodb.Client`: A client for interacting with AWS DynamoDB.
- `error`: Error message, if any.

## User Balance

### CheckUsersExist

```go
func CheckUsersExist(dbSvc *dynamodb.Client, accountIds []string) ([]string, error)
```

**Purpose:** Checks if users exist in the DynamoDB table.

**Parameters:**
- `dbSvc`: DynamoDB client.
- `accountIds`: List of user account IDs.

**Returns:**
- `[]string`: List of account IDs that do not exist.
- `error`: Error message if the operation fails.

### CreateAccountWithBalance

```go
func CreateAccountWithBalance(dbSvc *dynamodb.Client, accountId string, amount float64) error
```

**Purpose:** Creates a new account with an initial balance.

**Parameters:**
- `dbSvc`: DynamoDB client.
- `accountId`: The unique identifier for the account.
- `amount`: Initial amount to be set for the account.

**Returns:**
- `error`: Error message if the operation fails.

### InquireBalance

```go
func InquireBalance(dbSvc *dynamodb.Client, AccountID string) (float64, error)
```

**Purpose:** Inquires about the balance of a user's account.

**Parameters:**
- `dbSvc`: DynamoDB client.
- `AccountID`: The unique identifier for the account.

**Returns:**
- `float64`: The current balance of the account.
- `error`: Error message if the operation fails.

## Transactions

### TransferCredits

```go
func TransferCredits(dbSvc *dynamodb.Client, fromAccountID, toAccountID string, amount float64) error
```

**Purpose:** Transfers credits from one account to another.

**Parameters:**
- `dbSvc`: DynamoDB client.
- `fromAccountID`: The account ID to debit.
- `toAccountID`: The account ID to credit.
- `amount`: The amount to be transferred.

**Returns:**
- `error`: Error message if the operation fails.

### GetTransactions

```go
func GetTransactions(dbSvc *dynamodb.Client, accountID string, limit int32, lastTransactionID string) ([]LedgerEntry, string, error)
```

**Purpose:** Retrieves a list of transactions for an account.

**Parameters:**
- `dbSvc`: DynamoDB client.
- `accountID`: The account ID.
- `limit`: Maximum number of transactions to retrieve.
- `lastTransactionID`: ID of the last transaction retrieved (for pagination).

**Returns:**
- `[]LedgerEntry`: List of ledger entries.
- `string`: The ID of the last transaction retrieved.
- `error`: Error message if the operation fails.

## Notifications

### HandleDynamoDBStream

```go
func HandleDynamoDBStream(ctx context.Context, event events.DynamoDBEvent) error
```

**Purpose:** Handles DynamoDB stream events to send email notifications.

**Parameters:**
- `ctx`: Context.
- `event`: DynamoDB event.

**Returns:**
- `error`: Error message if the operation fails.

### SendEmail

```go
func SendEmail(sesSvc *ses.Client, msg Message) error
```

**Purpose:** Sends an email notification.

**Parameters:**
- `sesSvc`: AWS SES client.
- `msg`: Message containing email details.

**Returns:**
- `error`: Error message if the operation fails.

### SendSMS

```go
func SendSMS(sms SMS) error
```

**Purpose:** Sends an SMS notification.

**Parameters:**
- `sms`: SMS structure containing the details for the SMS message.

**Returns:**
- `error`: Error message if the operation fails.

## Roadmap for Planned Features

**Short-term Goals:**
1. Refine the current DynamoDB schema to improve performance for high transaction volumes.
2. Implement a caching layer to reduce read latency for frequently accessed data.
3. Develop a more robust error handling and logging mechanism for easier debugging and maintenance.

**Mid-term Goals:**
1. Integrate

 with additional AWS services for analytics and real-time monitoring of transactions.
2. Support multi-currency transactions and automatic currency conversion based on real-time exchange rates.
3. Provide a user interface for account management and transaction history.

**Long-term Goals:**
1. Expand the ledger system to support blockchain technologies for increased security and transparency.
2. Establish a plugin system allowing third-party extensions and integrations.
3. Implement AI-driven fraud detection and prevention systems.


