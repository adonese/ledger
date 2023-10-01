package ledger

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Balances struct {
	AccountID string  `json:"AccountID"`
	Amount    float64 `json:"Amount"`
	// add meta-fields here
}

// UserBalance represents the user's balance in the DynamoDB table
type UserBalance struct {
	AccountID string  `json:"AccountID"`
	Amount    float64 `json:"Amount"`
}

func CheckUserExists(dbSvc *dynamodb.DynamoDB, accountId string) error {
	// Prepare the input for the GetItem operation
	input := &dynamodb.GetItemInput{
		TableName: aws.String("UserBalanceTable"), // Replace with your DynamoDB table name
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {
				S: aws.String(accountId),
			},
		},
	}
	// Retrieve the item from DynamoDB
	result, err := dbSvc.GetItem(input)
	if err != nil {
		return err
	}

	if result.Item == nil {
		return errors.New("empty_user")
	} else {
		log.Printf("the found user is: %#v", result.Item)
		return nil
	}
}

func CreateAccountWithBalance(dbSvc *dynamodb.DynamoDB, accountId string, amount float64) error {
	item := map[string]*dynamodb.AttributeValue{
		"AccountID": {
			S: aws.String(accountId),
		},
		"Amount": {
			N: aws.String(fmt.Sprintf("%f", amount)),
		},
		"CreatedAt": {
			N: aws.String(fmt.Sprintf("%d", getCurrentTimestamp())),
		},
	}

	// Put the item into the DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName: aws.String("UserBalanceTable"),
		Item:      item,
	}

	_, err := dbSvc.PutItem(input)
	log.Printf("the error is: %v", err)
	return err
}

// Function to inquire about a user's balance
func InquireBalance(dbSvc *dynamodb.DynamoDB, AccountID string) (float64, error) {
	result, err := dbSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("UserBalanceTable"),
		Key: map[string]*dynamodb.AttributeValue{
			"AccountID": {S: aws.String(AccountID)},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to inquire balance for user %s: %v", AccountID, err)
	}
	if result.Item == nil {
		return 0, fmt.Errorf("user %s does not exist", AccountID)
	}
	userBalance := UserBalance{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &userBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal user balance for user %s: %v", AccountID, err)
	}
	return userBalance.Amount, nil
}

// Function to transfer credits from one user to another
func TransferCredits(dbSvc *dynamodb.DynamoDB, fromAccountID, toAccountID string, amount float64) error {
	// Create a new transaction input
	userBalance, err := InquireBalance(dbSvc, fromAccountID)
	if err != nil || amount > userBalance {
		return errors.New("insufficient balance")
	}
	debitEntry := LedgerEntry{
		AccountID:     fromAccountID,
		Amount:        amount,
		TransactionID: "1212",
		Type:          "debit",
	}
	creditEntry := LedgerEntry{
		AccountID:     toAccountID,
		Amount:        amount,
		TransactionID: "1212",
		Type:          "credit",
	}

	// Marshal the entry into a DynamoDB attribute value map
	avDebit, err := dynamodbattribute.MarshalMap(debitEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}
	avCredit, err := dynamodbattribute.MarshalMap(creditEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Update: &dynamodb.Update{
					TableName: aws.String("UserBalanceTable"),
					Key: map[string]*dynamodb.AttributeValue{
						"AccountID": {S: aws.String(fromAccountID)},
					},
					UpdateExpression:          aws.String("SET Amount = Amount - :amount"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":amount": {N: aws.String(fmt.Sprintf("%.2f", amount))}},
				},
			},
			{
				Update: &dynamodb.Update{
					TableName: aws.String("UserBalanceTable"),
					Key: map[string]*dynamodb.AttributeValue{
						"AccountID": {S: aws.String(toAccountID)}, // Replace with the other account's ID
					},
					UpdateExpression:          aws.String("SET Amount = Amount + :amount"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{":amount": {N: aws.String(fmt.Sprintf("%.2f", amount))}},
				},
			},
			{Put: &dynamodb.Put{
				TableName: aws.String("LedgerTable"),
				Item:      avDebit,
			}}, // PUT debit
			{Put: &dynamodb.Put{
				TableName: aws.String("LedgerTable"),
				Item:      avCredit,
			}}, // PUT credit
		},
	}

	// Perform the transaction
	_, err = dbSvc.TransactWriteItems(input)
	if err != nil {
		return fmt.Errorf("failed to debit from balance for user %s: %v", fromAccountID, err)
	}
	return nil
}

// Helper function to get the current timestamp
func getCurrentTimestamp() int64 {
	// Get the current time in UTC
	currentTime := time.Now().UTC()

	// Get the Unix timestamp (number of seconds since January 1, 1970)
	timestamp := currentTime.Unix()

	return timestamp
}
