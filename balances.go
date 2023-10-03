package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
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

func CheckUsersExist(dbSvc *dynamodb.Client, accountIds []string) ([]string, error) {
	// Prepare the input for the BatchGetItem operation
	keys := make([]map[string]types.AttributeValue, len(accountIds))
	var err error
	for i, accountId := range accountIds {
		keys[i] = map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: accountId},
		}
	}
	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			"UserBalanceTable": {
				Keys: keys,
			},
		},
	}

	// Retrieve the items from DynamoDB
	result, err := dbSvc.BatchGetItem(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var notFoundUsers []string
	var foundIds []string
	for _, item := range result.Responses["UserBalanceTable"] {
		if item != nil {
			foundIds = append(foundIds, item["AccountID"].(*types.AttributeValueMemberS).Value)
		}
	}

	for _, val := range accountIds {
		if !slices.Contains(foundIds, val) {
			notFoundUsers = append(notFoundUsers, val)
			err = errors.New("user_not_found")
		}
	}

	return notFoundUsers, err
}

func CreateAccountWithBalance(dbSvc *dynamodb.Client, accountId string, amount float64) error {
	item := map[string]types.AttributeValue{
		"AccountID": &types.AttributeValueMemberS{
			Value: accountId,
		},
		"Amount": &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%f", amount),
		},
		"CreatedAt": &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%d", getCurrentTimestamp()),
		},
	}

	// Put the item into the DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName: aws.String("UserBalanceTable"),
		Item:      item,
	}

	_, err := dbSvc.PutItem(context.TODO(), input)
	log.Printf("the error is: %v", err)
	return err
}

// Function to inquire about a user's balance
func InquireBalance(dbSvc *dynamodb.Client, AccountID string) (float64, error) {
	result, err := dbSvc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("UserBalanceTable"),
		Key: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: AccountID},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to inquire balance for user %s: %v", AccountID, err)
	}
	if result.Item == nil {
		return 0, fmt.Errorf("user %s does not exist", AccountID)
	}
	userBalance := UserBalance{}
	err = attributevalue.UnmarshalMap(result.Item, &userBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal user balance for user %s: %v", AccountID, err)
	}
	return userBalance.Amount, nil
}

func TransferCredits(dbSvc *dynamodb.Client, fromAccountID, toAccountID string, amount float64) error {
	// Create a new transaction input
	uid := uuid.New().String()
	userBalance, err := InquireBalance(dbSvc, fromAccountID)
	if err != nil || amount > userBalance {
		return errors.New("insufficient balance")
	}
	debitEntry := LedgerEntry{
		AccountID:     fromAccountID,
		Amount:        amount,
		TransactionID: uid,
		Type:          "debit",
	}
	creditEntry := LedgerEntry{
		AccountID:     toAccountID,
		Amount:        amount,
		TransactionID: uid,
		Type:          "credit",
	}

	// Marshal the entry into a DynamoDB attribute value map
	avDebit, err := attributevalue.MarshalMap(debitEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}
	avCredit, err := attributevalue.MarshalMap(creditEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String("UserBalanceTable"),
					Key: map[string]types.AttributeValue{
						"AccountID": &types.AttributeValueMemberS{Value: fromAccountID},
					},
					UpdateExpression:          aws.String("SET Amount = Amount - :amount"),
					ExpressionAttributeValues: map[string]types.AttributeValue{":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)}},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String("UserBalanceTable"),
					Key: map[string]types.AttributeValue{
						"AccountID": &types.AttributeValueMemberS{Value: toAccountID},
					},
					UpdateExpression:          aws.String("SET Amount = Amount + :amount"),
					ExpressionAttributeValues: map[string]types.AttributeValue{":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)}},
				},
			},
			{Put: &types.Put{
				TableName: aws.String("LedgerTable"),
				Item:      avDebit,
			}}, // PUT debit
			{Put: &types.Put{
				TableName: aws.String("LedgerTable"),
				Item:      avCredit,
			}}, // PUT credit
		},
	}

	// Perform the transaction
	_, err = dbSvc.TransactWriteItems(context.TODO(), input)
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
