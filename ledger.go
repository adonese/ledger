package ledger

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// InitializeLedger is a helper function to auth aws
func InitializeLedger(accessKey, secretKey, region string) (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	dyn := dynamodb.New(sess)
	return dyn, nil
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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
	})
	if err != nil {
		log.Fatal("Failed to create DynamoDB session:", err)
	}

	// Create a DynamoDB client
	dbSvc := dynamodb.New(sess)

	// Perform credit and debit transactions``
	err = RecordCredit(dbSvc, "account_id_1", 100.0)
	if err != nil {
		log.Fatal("Failed to record credit transaction:", err)
	}

	err = RecordDebit(dbSvc, "account_id_1", 50.0)
	if err != nil {
		log.Fatal("Failed to record debit transaction:", err)
	}
}

func RecordCredit(db *dynamodb.DynamoDB, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID: accountID,
		Amount:    amount,
		Type:      "credit",
	}

	// Marshal the entry into a DynamoDB attribute value map
	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      av,
	}

	// Put the ledger entry into the DynamoDB table
	_, err = db.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to record credit transaction: %v", err)
	}

	return nil
}

func RecordDebit(db *dynamodb.DynamoDB, accountID string, amount float64) error {
	// Create a new ledger entry
	entry := LedgerEntry{
		AccountID:     accountID,
		TransactionID: "1234",
		Amount:        amount,
		Type:          "debit",
	}

	// Marshal the entry into a DynamoDB attribute value map
	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	// Create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      av,
	}

	// Put the ledger entry into the DynamoDB table
	_, err = db.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to record debit transaction: %v", err)
	}

	return nil
}

// Function to store a transaction in the LedgerTable
func storeTransaction(dbSvc *dynamodb.DynamoDB, userID, transactionType string, amount float64) error {
	item := map[string]*dynamodb.AttributeValue{
		"AccountID":     {S: aws.String(userID)},
		"TransactionID": {S: aws.String("")},
		"Amount":        {N: aws.String(fmt.Sprintf("%.2f", amount))},
		"Type":          {S: aws.String(transactionType)},
		"Currency":      {S: aws.String("SDG")}, // Assuming a fixed currency for transactions
		"Timestamp":     {S: aws.String("")},
	}

	_, err := dbSvc.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("LedgerTable"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store transaction for user %s: %v", userID, err)
	}

	return nil
}
