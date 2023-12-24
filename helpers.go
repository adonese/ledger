package ledger

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// The StoreTransaction function stores the details of a transaction
func saveToTransactionTable(dbSvc *dynamodb.Client, transaction TransactionEntry, status int) error {
	transaction.Status = &status

	// Marshal the transaction into a DynamoDB attribute value map
	avTransaction, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction entry: %v", err)
	}

	// Define the DynamoDB transaction input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(TransactionsTable), // Replace with the actual name of your table
		Item:      avTransaction,
	}

	// Execute the transaction
	_, err = dbSvc.PutItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to store transaction: %v", err)
	}

	return nil
}
