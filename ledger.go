// Package ledger provides a set of functions to manage financial transactions
// and user balances in a ledger system. It supports operations like checking
// user existence, creating accounts, inquiring balances, transferring credits,
// and recording transactions. It uses AWS DynamoDB for data storage and AWS SES
// for sending notifications.

package ledger

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var AWS_REGION = "us-east-1"

// InitializeLedger initializes the DynamoDB client using AWS credentials.
// It takes an access key, a secret key, and a region and returns a DynamoDB client
// and an error if the initialization fails.
func InitializeLedger(accessKey, secretKey, region string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// NewS3 initializes an S3 client using AWS credentials.
// It takes an access key, a secret key, and a region and returns an S3 client
// and an error if the initialization fails.
func NewS3(accessKey, secretKey, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil

}

// LedgerEntry represents a single entry in the ledger table.
// It includes the account ID, transaction ID, the amount transacted,
// the type of transaction (debit or credit), and the time of transaction.
type LedgerEntry struct {
	AccountID           string  `dynamodbav:"AccountID" json:"account_id,omitempty"`
	SystemTransactionID string  `dynamodbav:"TransactionID" json:"transaction_id,omitempty"`
	Amount              float64 `dynamodbav:"Amount" json:"amount,omitempty"`
	Type                string  `dynamodbav:"Type" json:"type,omitempty"`
	Time                int64   `dynamodbav:"Time" json:"time,omitempty"`
	TenantID            string  `dynamodbav:"TenantID" json:"tenant_id,omitempty"`
	InitiatorUUID       string  `dynamodbav:"UUID" json:"uuid,omitempty"`
}
