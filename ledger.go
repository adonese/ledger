package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// InitializeLedger is a helper function to authenticate with AWS and create a DynamoDB client
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

// NewS3 returns a new S3 object
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

// LedgerEntry represents a single entry onto LedgerTable
type LedgerEntry struct {
	AccountID     string  `json:"AccountID"`
	TransactionID string  `json:"TransactionID"`
	Amount        float64 `json:"Amount"`
	Type          string  `json:"Type"`
}

func test() {
	// Create a new DynamoDB session
	var _dbSvc *dynamodb.Client

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-north-1"),
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

func RecordCredit(client *dynamodb.Client, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID: accountID,
		Amount:    amount,
		Type:      "credit",
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

func RecordDebit(client *dynamodb.Client, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID:     accountID,
		TransactionID: "1234",
		Amount:        amount,
		Type:          "debit",
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
		return fmt.Errorf("failed to record debit transaction: %v", err)
	}

	return nil
}

// Function to store a transaction in the LedgerTable
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
