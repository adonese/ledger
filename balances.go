package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/segmentio/ksuid"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
func CheckUsersExist(dbSvc *dynamodb.Client, tenantId string, accountIds []string) ([]string, error) {
	// Prepare the input for the BatchGetItem operation
	if tenantId == "" {
		tenantId = "nil"
	}
	keys := make([]map[string]types.AttributeValue, len(accountIds))
	var err error
	for i, accountId := range accountIds {
		keys[i] = map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: accountId},
			"TenantID":  &types.AttributeValueMemberS{Value: tenantId},
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
//
// FIXME(adonese): currently this creates a destructive operation where it overrides an existing user.
// the only way we're yet allowing this, is because the logic is managed via another indirection layer.
func CreateAccountWithBalance(dbSvc *dynamodb.Client, tenantId, accountId string, amount float64) error {
	if tenantId == "" {
		tenantId = "nil" // default value for old clients
	}
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
		"TenantID":            &types.AttributeValueMemberS{Value: tenantId},
	}

	conditionExpression := "attribute_not_exists(AccountID) AND attribute_not_exists(TenantID)"

	// Put the item into the DynamoDB table
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(NilUsers),
		Item:                item,
		ConditionExpression: &conditionExpression,
	}

	_, err := dbSvc.PutItem(context.TODO(), input)
	log.Printf("the error is: %v", err)
	return err
}

func CreateAccount(dbSvc *dynamodb.Client, tenantId string, user User) error {
	if tenantId == "" {
		tenantId = "nil"
	}
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
		"TenantID":            &types.AttributeValueMemberS{Value: tenantId},
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

// GetAccount retrieves an account by tenant ID and account ID.
func GetAccount(ctx context.Context, dbSvc *dynamodb.Client, trEntry TransactionEntry) (*User, error) {
	if trEntry.TenantID == "" {
		trEntry.TenantID = "nil"
	}
	key := map[string]types.AttributeValue{
		"TenantID":  &types.AttributeValueMemberS{Value: trEntry.TenantID},
		"AccountID": &types.AttributeValueMemberS{Value: trEntry.AccountID},
	}

	result, err := dbSvc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("NilUsers"),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, errors.New("uncaught error: empty user!")
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %v", err)
	}

	return &user, nil
}

// InquireBalance inquires the balance of a given user account.
// It takes a DynamoDB client and an account ID, returning the balance
// as a float64 and an error if the inquiry fails or the user does not exist.
func InquireBalance(dbSvc *dynamodb.Client, tenantId, AccountID string) (float64, error) {
	if tenantId == "" {
		tenantId = "nil"
	}
	result, err := dbSvc.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(NilUsers),
		Key: map[string]types.AttributeValue{
			"AccountID": &types.AttributeValueMemberS{Value: AccountID},
			"TenantID":  &types.AttributeValueMemberS{Value: tenantId},
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
func TransferCredits(dbSvc *dynamodb.Client, trEntry TransactionEntry) error {
	if trEntry.AccountID == "" {
		return errors.New("you must provide Account ID, substitute it for FromAccount to mimic the older api")
	}
	if trEntry.TenantID == "" {
		trEntry.TenantID = "nil"
	}
	timestamp := getCurrentTimestamp()
	var transactionStatus int = 1
	// We are using ksuid in order to have a sortable randomized UUIDs with great entropy
	uid := ksuid.New().String()

	// Define the transaction
	transaction := TransactionEntry{
		TenantID:        trEntry.TenantID,
		AccountID:       trEntry.FromAccount,
		TransactionID:   uid,
		FromAccount:     trEntry.FromAccount,
		ToAccount:       trEntry.ToAccount,
		Amount:          trEntry.Amount,
		Comment:         "Transfer credits",
		TransactionDate: timestamp,
		Status:          &transactionStatus,
	}

	user, err := GetAccount(context.TODO(), dbSvc, trEntry)
	if err != nil || user == nil {
		SaveToTransactionTable(dbSvc, trEntry.TenantID, transaction, transactionStatus)
		return fmt.Errorf("error in retrieving user: %v", err)
	}

	if trEntry.Amount > user.Amount {
		SaveToTransactionTable(dbSvc, trEntry.TenantID, transaction, transactionStatus)
		return errors.New("insufficient balance")
	}

	debitEntry := LedgerEntry{
		TenantID:      trEntry.TenantID,
		AccountID:     trEntry.FromAccount,
		Amount:        trEntry.Amount,
		TransactionID: uid,
		Type:          "debit",
		Time:          timestamp,
	}
	creditEntry := LedgerEntry{
		TenantID:      trEntry.TenantID,
		AccountID:     trEntry.ToAccount,
		Amount:        trEntry.Amount,
		TransactionID: uid,
		Type:          "credit",
		Time:          timestamp,
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

	// Perform the debit transaction
	debitInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"TenantID":  &types.AttributeValueMemberS{Value: trEntry.TenantID},
						"AccountID": &types.AttributeValueMemberS{Value: trEntry.FromAccount},
					},
					UpdateExpression:    aws.String("SET amount = amount - :amount, Version = :newVersion"),
					ConditionExpression: aws.String("attribute_not_exists(Version) OR Version = :oldVersion"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
						":oldVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(user.Version, 10)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
					},
				},
			},
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avDebit,
			}}, // PUT debit
		},
	}

	_, err = dbSvc.TransactWriteItems(context.TODO(), debitInput)
	if err != nil {
		transactionStatus = 1
		if err := SaveToTransactionTable(dbSvc, trEntry.TenantID, transaction, transactionStatus); err != nil {
			panic(err)
		}
		return fmt.Errorf("failed to debit from balance for user %s: %v", trEntry.FromAccount, err)
	}

	// Perform the credit transaction
	creditInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"TenantID":  &types.AttributeValueMemberS{Value: trEntry.TenantID},
						"AccountID": &types.AttributeValueMemberS{Value: trEntry.ToAccount},
					},
					UpdateExpression:    aws.String("SET amount = amount + :amount, Version = :newVersion"),
					ConditionExpression: aws.String("attribute_exists(AccountID) AND TenantID = :tenantID"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
						":tenantID":   &types.AttributeValueMemberS{Value: trEntry.TenantID},
					},
				},
			},
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avCredit,
			}},
		},
	}

	_, err = dbSvc.TransactWriteItems(context.TODO(), creditInput)
	if err != nil {
		// Rollback debit if credit fails
		rollbackInput := &dynamodb.UpdateItemInput{
			TableName: aws.String(NilUsers),
			Key: map[string]types.AttributeValue{
				"TenantID":  &types.AttributeValueMemberS{Value: trEntry.TenantID},
				"AccountID": &types.AttributeValueMemberS{Value: trEntry.FromAccount},
			},
			UpdateExpression:    aws.String("SET amount = amount + :amount, Version = :newVersion"),
			ConditionExpression: aws.String("attribute_not_exists(Version) OR Version = :oldVersion"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
				":oldVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(user.Version, 10)},
				":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
			},
		}

		_, rollbackErr := dbSvc.UpdateItem(context.TODO(), rollbackInput)
		if rollbackErr != nil {
			panic(fmt.Errorf("failed to rollback debit for user %s: %v", trEntry.FromAccount, rollbackErr))
		}

		transactionStatus = 1
		if err := SaveToTransactionTable(dbSvc, trEntry.TenantID, transaction, transactionStatus); err != nil {
			panic(err)
		}
		return fmt.Errorf("failed to credit to balance for user %s: %v", trEntry.ToAccount, err)
	}

	transactionStatus = 0
	if err := SaveToTransactionTable(dbSvc, trEntry.TenantID, transaction, transactionStatus); err != nil {
		panic(err)
	}

	return nil
}

// GetTransactions retrieves a list of transactions for a specified tenant and account.
// It takes a DynamoDB client, a tenant ID, an account ID, a limit for the number of transactions
// to retrieve, and an optional lastTransactionID for pagination.
// It returns a slice of LedgerEntry, the ID of the last transaction, and an error, if any.
func GetTransactions(dbSvc *dynamodb.Client, tenantID, accountID string, limit int32, lastTransactionID string) ([]LedgerEntry, string, error) {
	if tenantID == "" {
		tenantID = "nil"
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String("TransactionsTable"),
		KeyConditionExpression: aws.String("TenantID = :tenantId AND AccountID = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tenantId":  &types.AttributeValueMemberS{Value: tenantID},
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit: aws.Int32(limit),
	}

	// If a lastTransactionID was provided, include it in the input
	if lastTransactionID != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"TenantID":      &types.AttributeValueMemberS{Value: tenantID},
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

// GetDetailedTransactions retrieves a list of transactions for a specified tenant and account.
// It takes a DynamoDB client, a tenant ID, an account ID, and a limit for the number of transactions
// to retrieve. It returns a slice of TransactionEntry and an error, if any.
func GetDetailedTransactions(dbSvc *dynamodb.Client, tenantID, accountID string, limit int32) ([]TransactionEntry, error) {
	// Query for transactions sent by the account
	if tenantID == "" {
		tenantID = "nil"
	}
	sentTransactions, _, err := getTransactionsByIndex(dbSvc, tenantID, "FromAccountIndex", "FromAccount", accountID, limit, "")
	if err != nil {
		return nil, err
	}
	// Query for transactions received by the account
	receivedTransactions, _, err := getTransactionsByIndex(dbSvc, tenantID, "ToAccountIndex", "ToAccount", accountID, limit, "")
	if err != nil {
		return nil, err
	}

	// Combine the transactions into a single list
	allTransactions := append(sentTransactions, receivedTransactions...)

	return allTransactions, nil
}

// getTransactionsByIndex is a helper function that queries for transactions on a specific index.
func getTransactionsByIndex(dbSvc *dynamodb.Client, tenantID, indexName, attributeName, accountID string, limit int32, lastTransactionID string) ([]TransactionEntry, string, error) {
	if tenantID == "" {
		tenantID = "nil"
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String("TransactionsTable"),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("TenantID = :tenantId AND " + attributeName + " = :accountId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tenantId":  &types.AttributeValueMemberS{Value: tenantID},
			":accountId": &types.AttributeValueMemberS{Value: accountID},
		},
		Limit:            aws.Int32(limit),
		ScanIndexForward: aws.Bool(false),
	}

	if lastTransactionID != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"TenantID":      &types.AttributeValueMemberS{Value: tenantID},
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

func GetAllNilTransactions(ctx context.Context, dbSvc *dynamodb.Client, tenantId string, filter TransactionFilter) ([]TransactionEntry, map[string]types.AttributeValue, error) {

	if tenantId == "" {
		tenantId = "nil"
	}
	var filterExpression strings.Builder
	expressionAttributeValues := map[string]types.AttributeValue{}

	// Add TenantID to filter expression
	filterExpression.WriteString("TenantID = :tenantId")
	expressionAttributeValues[":tenantId"] = &types.AttributeValueMemberS{Value: tenantId}

	// Building filter expressions
	if filter.AccountID != "" {
		filterExpression.WriteString(" AND (FromAccount = :accountID OR ToAccount = :accountID)")
		expressionAttributeValues[":accountID"] = &types.AttributeValueMemberS{Value: filter.AccountID}
	}

	if filter.TransactionStatus != nil {
		filterExpression.WriteString(" AND TransactionStatus = :transactionStatus")
		expressionAttributeValues[":transactionStatus"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*filter.TransactionStatus)}
	}

	if filter.StartTime != 0 && filter.EndTime != 0 {
		filterExpression.WriteString(" AND TransactionDate BETWEEN :startTime AND :endTime")
		expressionAttributeValues[":startTime"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(filter.StartTime, 10)}
		expressionAttributeValues[":endTime"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(filter.EndTime, 10)}
	}

	if filter.Limit == 0 {
		filter.Limit = 25
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String("TransactionsTable"),
		Limit:                     aws.Int32(filter.Limit),
		FilterExpression:          aws.String(filterExpression.String()),
		ExpressionAttributeValues: expressionAttributeValues,
	}

	if len(filter.LastEvaluatedKey) > 0 {
		scanInput.ExclusiveStartKey = filter.LastEvaluatedKey
	}

	output, err := dbSvc.Scan(ctx, scanInput)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	var transactions []TransactionEntry
	err = attributevalue.UnmarshalListOfMaps(output.Items, &transactions)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal transactions: %v", err)
	}

	return transactions, output.LastEvaluatedKey, nil
}

// Helper function to append filter expressions
func addFilterExpression(existing, add string) string {
	if existing != "" {
		return existing + " AND " + add
	}
	return add
}

// Helper function to get the current timestamp
func getCurrentTimestamp() int64 {
	// Get the current time in UTC
	currentTime := time.Now().UTC()

	// Get the Unix timestamp (number of seconds since January 1, 1970)
	timestamp := currentTime.Unix()

	return timestamp
}
