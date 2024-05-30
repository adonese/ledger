// Package ledger provides a set of functions to manage financial transactions
// and user balances in a ledger system. It supports operations like checking
// user existence, creating accounts, inquiring balances, transferring credits,
// and recording transactions. It uses AWS DynamoDB for data storage and AWS SES
// for sending notifications.

package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	AccountID     string  `dynamodbav:"AccountID" json:"account_id,omitempty"`
	TransactionID string  `dynamodbav:"TransactionID" json:"transaction_id,omitempty"`
	Amount        float64 `dynamodbav:"Amount" json:"amount,omitempty"`
	Type          string  `dynamodbav:"Type" json:"type,omitempty"`
	Time          int64   `dynamodbav:"Time" json:"time,omitempty"`
	TenantID      string  `dynamodbav:"TenantID" json:"tenant_id,omitempty"`
}

func test() {
	// Create a new DynamoDB session
	var _dbSvc *dynamodb.Client

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(AWS_REGION),
	)
	if err != nil {
		log.Fatal("Failed to create DynamoDB session:", err)
	}

	_dbSvc = dynamodb.NewFromConfig(cfg)

	// Perform credit and debit transactions``
	err = RecordCredit(_dbSvc, "account_id_1", 100.0)
	if err != nil {
		log.Fatal("Failed to record credit transaction:", err)
	}

	err = RecordDebit(_dbSvc, "account_id_1", 50.0)
	if err != nil {
		log.Fatal("Failed to record debit transaction:", err)
	}
}

// RecordCredit records a credit transaction for an account.
// It takes a DynamoDB client, an account ID, and the amount to be credited.
// It returns an error if the recording fails.
func RecordCredit(client *dynamodb.Client, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID:     accountID,
		Amount:        amount,
		Type:          "credit",
		TransactionID: uuid.NewString(),
		Time:          getCurrentTimestamp(),
	}

	// Marshal the entry into a DynamoDB attribute value map
	av, err := attributevalue.MarshalMap(entry)
	if err != nil {
		return errors.New("failed to marshal ledger entry")
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      av,
	}

	// Put the ledger entry into the DynamoDB table
	_, err = client.PutItem(context.TODO(), input)
	if err != nil {
		return errors.New("failed to record credit transaction")
	}

	return nil
}

// RecordDebit records a debit transaction for an account.
// It takes a DynamoDB client, an account ID, and the amount to be debited.
// It returns an error if the recording fails.
func RecordDebit(client *dynamodb.Client, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID:     accountID,
		TransactionID: uuid.NewString(),
		Amount:        amount,
		Type:          "debit",
		Time:          getCurrentTimestamp(),
	}

	// Marshal the entry into a DynamoDB attribute value map
	av, err := attributevalue.MarshalMap(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      av,
	}

	// Put the ledger entry into the DynamoDB table
	_, err = client.PutItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to record debit transaction: %v", err)
	}

	return nil
}

// storeTransaction stores a transaction in the ledger table.
// It takes a DynamoDB client, a user ID, the type of transaction, and the amount.
// It returns an error if the transaction cannot be stored.
func storeTransaction(dbSvc *dynamodb.Client, userID, transactionType string, amount float64) error {
	item := map[string]types.AttributeValue{
		"AccountID":     &types.AttributeValueMemberS{Value: userID},
		"TransactionID": &types.AttributeValueMemberS{Value: ""},
		"Amount":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)},
		"Type":          &types.AttributeValueMemberS{Value: transactionType},
		"Currency":      &types.AttributeValueMemberS{Value: "SDG"},
		"Timestamp":     &types.AttributeValueMemberS{Value: ""},
	}

	_, err := dbSvc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store transaction for user %s: %v", userID, err)
	}

	return nil
}
