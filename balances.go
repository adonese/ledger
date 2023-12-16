package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var (
	NilUsers          = "NilUsers"
	LedgerTable       = "LedgerTable"
	TransactionsTable = "TransactionsTable"
)

// Balances represents the amount of money in a user's account.
// AccountID is a unique identifier for the account, and Amount
// is the balance available in the account.
type Balances struct {
	AccountID string  `json:"AccountID"`
	Amount    float64 `json:"Amount"`
	// add meta-fields here
}

// UserBalance represents the user's balance in the DynamoDB table.
// It includes the AccountID and the associated Amount.
type UserBalance struct {
	AccountID string  `json:"AccountID"`
	Amount    float64 `json:"Amount"`
}

// CheckUsersExist checks if the provided account IDs exist in the DynamoDB table.
// It takes a DynamoDB client and a slice of account IDs and returns a slice of
// non-existent account IDs and an error, if any.
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
			NilUsers: {
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
	for _, item := range result.Responses[NilUsers] {
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

// CreateAccountWithBalance creates a new user account with an initial balance.
// It takes a DynamoDB client, an account ID, and an amount to be set as the initial
// balance. It returns an error if the account creation fails.
func CreateAccountWithBalance(dbSvc *dynamodb.Client, accountId string, amount float64) error {
	// parse float to string

	item := map[string]types.AttributeValue{
		"AccountID":           &types.AttributeValueMemberS{Value: accountId},
		"full_name":           &types.AttributeValueMemberS{Value: "test-account"},
		"birthday":            &types.AttributeValueMemberS{Value: ""},
		"city":                &types.AttributeValueMemberS{Value: ""},
		"dependants":          &types.AttributeValueMemberN{Value: "0"},
		"income_last_year":    &types.AttributeValueMemberN{Value: "0"},
		"enroll_smes_program": &types.AttributeValueMemberBOOL{Value: false},
		"confirm":             &types.AttributeValueMemberBOOL{Value: false},
		"external_auth":       &types.AttributeValueMemberBOOL{Value: false},
		"password":            &types.AttributeValueMemberS{Value: ""},
		"created_at":          &types.AttributeValueMemberS{Value: time.Now().Local().String()},
		"is_verified":         &types.AttributeValueMemberBOOL{Value: true},
		"id_type":             &types.AttributeValueMemberS{Value: ""},
		"mobile_number":       &types.AttributeValueMemberS{Value: ""},
		"id_number":           &types.AttributeValueMemberS{Value: ""},
		"pic_id_card":         &types.AttributeValueMemberS{Value: ""},
		"amount":              &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)},
		"currency":            &types.AttributeValueMemberS{Value: "SDG"},
		"Version":             &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
	}

	// Put the item into the DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName: aws.String(NilUsers),
		Item:      item,
	}

	_, err := dbSvc.PutItem(context.TODO(), input)
	log.Printf("the error is: %v", err)
	return err
}

func CreateAccount(dbSvc *dynamodb.Client, user User) error {
	item := map[string]types.AttributeValue{
		"AccountID":           &types.AttributeValueMemberS{Value: user.AccountID},
		"full_name":           &types.AttributeValueMemberS{Value: user.FullName},
		"birthday":            &types.AttributeValueMemberS{Value: user.Birthday},
		"city":                &types.AttributeValueMemberS{Value: user.City},
		"dependants":          &types.AttributeValueMemberN{Value: strconv.Itoa(user.Dependants)},
		"income_last_year":    &types.AttributeValueMemberN{Value: strconv.Itoa(int(user.IncomeLastYear))},
		"enroll_smes_program": &types.AttributeValueMemberBOOL{Value: user.EnrollSMEsProgram},
		"confirm":             &types.AttributeValueMemberBOOL{Value: user.Confirm},
		"external_auth":       &types.AttributeValueMemberBOOL{Value: user.ExternalAuth},
		"password":            &types.AttributeValueMemberS{Value: user.Password},
		"created_at":          &types.AttributeValueMemberS{Value: time.Now().Local().String()},
		"is_verified":         &types.AttributeValueMemberBOOL{Value: user.IsVerified},
		"id_type":             &types.AttributeValueMemberS{Value: user.IDType},
		"mobile_number":       &types.AttributeValueMemberS{Value: user.MobileNumber},
		"id_number":           &types.AttributeValueMemberS{Value: user.IDNumber},
		"pic_id_card":         &types.AttributeValueMemberS{Value: user.PicIDCard},
		"amount":              &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", user.Amount)},
		"currency":            &types.AttributeValueMemberS{Value: "SDG"},
		"Version":             &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
	}

	// Put the item into the DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName: aws.String(NilUsers),
		Item:      item,
	}

	_, err := dbSvc.PutItem(context.TODO(), input)
	log.Printf("the error is: %v", err)
	return err
}

func GetAccount(ctx context.Context, dbSvc *dynamodb.Client, accountID string) (*User, error) {
	key := map[string]types.AttributeValue{
		"AccountID": &types.AttributeValueMemberS{Value: accountID},
	}

	out, err := dbSvc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(NilUsers),
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %v", err)
	}

	if out.Item == nil {
		return nil, fmt.Errorf("account not found: %s", accountID)
	}

	var account User
	err = attributevalue.UnmarshalMap(out.Item, &account)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return &account, nil
}

// InquireBalance inquires the balance of a given user account.
// It takes a DynamoDB client and an account ID, returning the balance
// as a float64 and an error if the inquiry fails or the user does not exist.
func InquireBalance(dbSvc *dynamodb.Client, AccountID string) (float64, error) {
	result, err := dbSvc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(NilUsers),
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

// TransferCredits transfers a specified amount from one account to another.
// It performs a transaction that debits one account and credits another.
// It takes a DynamoDB client, the account IDs for the sender and receiver, and
// the amount to transfer. It returns an error if the transfer fails due to
// insufficient funds or other issues.
func TransferCredits(dbSvc *dynamodb.Client, fromAccountID, toAccountID string, amount float64) error {
	// Create a new transaction input
	uid := uuid.New().String()
	user, err := GetAccount(context.TODO(), dbSvc, fromAccountID)
	if err != nil || user == nil {
		return fmt.Errorf("error in retrieving user: %v", err)
	}
	if amount > user.Amount {
		return errors.New("insufficient balance")
	}
	timestamp := getCurrentTimestamp()
	debitEntry := LedgerEntry{
		AccountID:     fromAccountID,
		Amount:        amount,
		TransactionID: uid,
		Type:          "debit",
		Time:          timestamp,
	}
	creditEntry := LedgerEntry{
		AccountID:     toAccountID,
		Amount:        amount,
		TransactionID: uid,
		Type:          "credit",
		Time:          timestamp,
	}

	// Define the transaction.
	transaction := TransactionEntry{
		AccountID:       fromAccountID,
		TransactionID:   uid,
		FromAccount:     fromAccountID,
		ToAccount:       toAccountID,
		Amount:          amount,
		Comment:         "Transfer credits", // Replace with actual comment
		TransactionDate: timestamp,
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
	// Marshal the transaction into a DynamoDB attribute value map.
	avTransaction, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction entry: %v", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"AccountID": &types.AttributeValueMemberS{Value: fromAccountID},
					},
					UpdateExpression:    aws.String("SET amount = amount - :amount, Version = :newVersion"),
					ConditionExpression: aws.String("attribute_not_exists(Version) OR Version = :oldVersion"),
					ExpressionAttributeValues: map[string]types.AttributeValue{":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)},
						":oldVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(user.Version, 10)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)}},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"AccountID": &types.AttributeValueMemberS{Value: toAccountID},
					},
					UpdateExpression: aws.String("SET amount = amount + :amount, Version = :newVersion"),
					ExpressionAttributeValues: map[string]types.AttributeValue{":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", amount)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)}},
				},
			},
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avDebit,
			}}, // PUT debit
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avCredit,
			}}, // PUT credit
			{
				Put: &types.Put{
					TableName: aws.String(TransactionsTable), // Replace with the actual name of your table
					Item:      avTransaction,
				},
			}, // put transaction

		},
	}

	// Perform the transaction
	_, err = dbSvc.TransactWriteItems(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to debit from balance for user %s: %v", fromAccountID, err)
	}
	return nil
}

// GetTransactions retrieves a list of transactions for a specified account.
// It takes a DynamoDB client, an account ID, a limit for the number of transactions
// to retrieve, and an optional lastTransactionID for pagination.
// It returns a slice of LedgerEntry, the ID of the last transaction, and an error, if any.
func GetTransactions(dbSvc *dynamodb.Client, accountID string, limit int32, lastTransactionID string) ([]LedgerEntry, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("LedgerTable"),
		KeyConditionExpression: aws.String("AccountID = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit: aws.Int32(limit),
	}

	// If a lastTransactionID was provided, include it in the input
	if lastTransactionID != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"AccountID":     &types.AttributeValueMemberS{Value: accountID},
			"TransactionID": &types.AttributeValueMemberS{Value: lastTransactionID},
		}
	}

	// Execute the query
	resp, err := dbSvc.Query(context.TODO(), input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch transactions: %v", err)
	}

	// Unmarshal the items
	var transactions []LedgerEntry
	err = attributevalue.UnmarshalListOfMaps(resp.Items, &transactions)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal transactions: %v", err)
	}

	// If there are more items to be fetched, return the TransactionID of the last item
	var newLastTransactionID string
	if resp.LastEvaluatedKey != nil {
		newLastTransactionID = resp.LastEvaluatedKey["TransactionID"].(*types.AttributeValueMemberS).Value
	}

	return transactions, newLastTransactionID, nil
}

// GetTransactions retrieves a list of transactions for a specified account.
// It takes a DynamoDB client, an account ID, and a limit for the number of transactions
// to retrieve. It returns a slice of TransactionEntry and an error, if any.
func GetDetailedTransactions(dbSvc *dynamodb.Client, accountID string, limit int32) ([]TransactionEntry, error) {
	// Query for transactions sent by the account
	sentTransactions, _, err := getTransactionsByIndex(dbSvc, "FromAccountIndex", "FromAccount", accountID, limit, "")
	if err != nil {
		return nil, err
	}

	// Query for transactions received by the account
	receivedTransactions, _, err := getTransactionsByIndex(dbSvc, "ToAccountIndex", "ToAccount", accountID, limit, "")
	if err != nil {
		return nil, err
	}

	// Combine the transactions into a single list
	allTransactions := append(sentTransactions, receivedTransactions...)

	return allTransactions, nil
}

// getTransactionsByIndex is a helper function that queries for transactions on a specific index.
func getTransactionsByIndex(dbSvc *dynamodb.Client, indexName string, attributeName string, accountID string, limit int32, lastTransactionID string) ([]TransactionEntry, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              &TransactionsTable,
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :accountId", attributeName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit: aws.Int32(limit),
	}

	if lastTransactionID != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"AccountID":     &types.AttributeValueMemberS{Value: accountID},
			"TransactionID": &types.AttributeValueMemberS{Value: lastTransactionID},
		}
	}

	resp, err := dbSvc.Query(context.TODO(), input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch transactions: %v", err)
	}

	var transactions []TransactionEntry
	err = attributevalue.UnmarshalListOfMaps(resp.Items, &transactions)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal transactions: %v", err)
	}

	var newLastTransactionID string
	if resp.LastEvaluatedKey != nil {
		newLastTransactionID = resp.LastEvaluatedKey["TransactionID"].(*types.AttributeValueMemberS).Value
	}

	return transactions, newLastTransactionID, nil
}

// Helper function to get the current timestamp
func getCurrentTimestamp() int64 {
	// Get the current time in UTC
	currentTime := time.Now().UTC()

	// Get the Unix timestamp (number of seconds since January 1, 1970)
	timestamp := currentTime.Unix()

	return timestamp
}
