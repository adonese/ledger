package ledger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/segmentio/ksuid"
)

// Optional: Implement MarshalDynamoDBAttributeValue for consistency
// func (s Status) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
// 	str, ok := statusEnumToString[s]
// 	if !ok {
// 		return nil, fmt.Errorf("unknown status: %d", s)
// 	}

// 	return &types.AttributeValueMemberS{Value: str}, nil
// }

const EscrowTransactionsTable = "EscrowTransactions"
const ESCROW_ACCOUNT = "NIL_ESCROW_ACCOUNT"
const ESCROW_TENANT = "ESCROW_TENANT"
const ServiceProvidersTransactions = "ServiceProviderTransactions"

func EscrowRequest(context context.Context, dbSvc *dynamodb.Client, esEntry EscrowEntry) (NilResponse, error) {
	var response NilResponse

	timestamp := getCurrentTimestamp()
	transactionStatus := StatusPending

	uid := ksuid.New().String()

	es := EscrowTransaction{
		FromAccount:   esEntry.FromAccount,
		FromTenantID:  esEntry.FromTenantID,
		ToAccount:     ESCROW_ACCOUNT, // We need to make sure that ESCROW_ACCOUNT is an exception
		ToTenantID:    ESCROW_TENANT,
		Amount:        esEntry.Amount,
		InitiatorUUID: esEntry.InitiatorUUID,
	}

	if _, err := EscrowTransferCredits(context, dbSvc, es); err != nil {
		return NilResponse{}, err
	}

	es.Status = transactionStatus
	es.SystemTransactionID = uid
	es.TransactionDate = timestamp

	// HERE we are supposed to ensure that from and to are actually matches what we want
	// Now, after you have done that, you should write these to Table
	// What we need here is a fully fledged EscrowTransaction
	cashOutProvider := "nil"
	if esEntry.CashoutProvider != "" {
		cashOutProvider = esEntry.CashoutProvider
	}
	esTransaction := EscrowTransaction{
		FromAccount:         esEntry.FromAccount,
		ToAccount:           esEntry.ToAccount,
		FromTenantID:        esEntry.FromTenantID,
		ToTenantID:          esEntry.ToTenantID,
		Amount:              esEntry.Amount,
		InitiatorUUID:       es.InitiatorUUID,
		SystemTransactionID: uid,
		TransactionDate:     timestamp,
		Timestamp:           getCurrentTimeZone(),
		Status:              StatusInProgress,
		Beneficiary:         esEntry.Beneficiary,
		TransientAccount:    ESCROW_ACCOUNT,
		TransientTenant:     ESCROW_TENANT,
		CashoutProvider:     cashOutProvider,
		ServiceProvider:     esEntry.ServiceProvider,
	}

	item, err := attributevalue.MarshalMap(esTransaction)
	if err != nil {
		// reverse the transfer here if fails!
		return NilResponse{}, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(EscrowTransactionsTable), // save it in escrow transactions table
		Item:      item,
	}

	if _, err := dbSvc.PutItem(context, input); err != nil {
		spew.Dump(item)
		// reverse the transfer here if fails!
		return NilResponse{}, fmt.Errorf("failed to put item into DynamoDB: %w - the payload is: %+v", err, item)
	}

	return response, nil
}

func EscrowTransferCredits(context context.Context, dbSvc *dynamodb.Client, trEntry EscrowTransaction) (NilResponse, error) {
	var response NilResponse
	if trEntry.FromAccount == "" || trEntry.ToAccount == "" {
		return response, errors.New("you must provide Account ID for both to/from account, substitute it for FromAccount to mimic the older api")
	}

	timestamp := getCurrentTimestamp()
	var transactionStatus int = 1
	uid := ksuid.New().String()

	combinedTenants := trEntry.FromTenantID + ":" + trEntry.ToTenantID

	transaction := TransactionEntry{
		TenantID:            combinedTenants,
		AccountID:           trEntry.FromAccount,
		SystemTransactionID: uid,
		FromAccount:         trEntry.FromAccount,
		ToAccount:           trEntry.ToAccount,
		Amount:              trEntry.Amount,
		Comment:             "Transfer credits",
		TransactionDate:     timestamp,
		Status:              &transactionStatus,
		InitiatorUUID:       trEntry.InitiatorUUID,
	}

	// Fetch sender account - sender here is the escrow account
	sender, err := GetAccount(context, dbSvc, TransactionEntry{AccountID: trEntry.FromAccount, FromAccount: trEntry.FromAccount, TenantID: trEntry.FromTenantID})
	if err != nil || sender == nil {
		SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus)
		response = NilResponse{
			Status:    "error",
			Code:      "user_not_found",
			Message:   "Error in retrieving sender.",
			Details:   fmt.Sprintf("Error in retrieving sender: %v", err),
			Timestamp: trEntry.Timestamp,
			Data: data{
				UUID:       trEntry.InitiatorUUID,
				SignedUUID: trEntry.SignedUUID,
			},
		}
		return response, err
	}

	// here it might be the case where by the output is for bankak or any other provider..
	if trEntry.CashoutProvider == "bok" {
		receiver, err := GetAccount(context, dbSvc, TransactionEntry{AccountID: trEntry.ToAccount, FromAccount: trEntry.ToAccount, TenantID: trEntry.ToTenantID})
		if err != nil || receiver == nil {
			SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus)
			response = NilResponse{
				Status:    "error",
				Code:      "user_not_found",
				Message:   "Error in retrieving receiver.",
				Details:   fmt.Sprintf("Error in retrieving receiver: %v", err),
				Timestamp: trEntry.Timestamp,
				Data: data{
					UUID:       trEntry.InitiatorUUID,
					SignedUUID: trEntry.SignedUUID,
				},
			}
			return response, err
		}

		if trEntry.Amount > sender.Amount {
			SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus)
			response = NilResponse{
				Status:    "error",
				Code:      "insufficient_balance",
				Message:   "Insufficient balance to complete the transaction.",
				Details:   "The sender does not have enough balance in their account.",
				Timestamp: trEntry.Timestamp,
				Data: data{
					UUID:       trEntry.InitiatorUUID,
					SignedUUID: trEntry.SignedUUID,
				},
			}
			return response, errors.New("insufficient balance")
		}
	}

	debitEntry := LedgerEntry{
		TenantID:            trEntry.FromTenantID,
		AccountID:           trEntry.FromAccount,
		Amount:              trEntry.Amount,
		SystemTransactionID: uid,
		Type:                "debit",
		Time:                timestamp,
		InitiatorUUID:       trEntry.InitiatorUUID,
	}
	// FIXME(adonese): if the cashout provider is bok, then the receiver is the escrow account for nilbok
	creditEntry := LedgerEntry{
		TenantID:            trEntry.ToTenantID,
		AccountID:           trEntry.ToAccount,
		Amount:              trEntry.Amount,
		SystemTransactionID: uid,
		Type:                "credit",
		Time:                timestamp,
		InitiatorUUID:       trEntry.InitiatorUUID,
	}

	avDebit, err := attributevalue.MarshalMap(debitEntry)
	if err != nil {
		return response, fmt.Errorf("failed to marshal ledger entry: %v", err)
	}
	avCredit, err := attributevalue.MarshalMap(creditEntry)
	if err != nil {
		return response, fmt.Errorf("failed to marshal ledger entry: %v", err)
	}

	debitInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"TenantID":  &types.AttributeValueMemberS{Value: trEntry.FromTenantID}, // use old tenant you got
						"AccountID": &types.AttributeValueMemberS{Value: trEntry.FromAccount},
					},
					UpdateExpression:    aws.String("SET amount = amount - :amount, Version = :newVersion"),
					ConditionExpression: aws.String("attribute_not_exists(Version) OR Version = :oldVersion"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
						":oldVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(sender.Version, 10)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
					},
				},
			},
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avDebit,
			}},
		},
	}

	_, err = dbSvc.TransactWriteItems(context, debitInput)
	if err != nil {
		transactionStatus = 1
		if err := SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus); err != nil {
			panic(err)
		}
		response = NilResponse{
			Status:    "error",
			Code:      "debit_failed",
			Message:   fmt.Sprintf("Failed to debit from balance for user %s", trEntry.FromAccount),
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: trEntry.Timestamp,
			Data: data{
				UUID:       trEntry.InitiatorUUID,
				SignedUUID: trEntry.SignedUUID,
			},
		}
		return response, fmt.Errorf("failed to debit from balance for user %s: %v", trEntry.FromAccount, err)
	}

	creditInput := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(NilUsers),
					Key: map[string]types.AttributeValue{
						"TenantID":  &types.AttributeValueMemberS{Value: trEntry.ToTenantID},
						"AccountID": &types.AttributeValueMemberS{Value: trEntry.ToAccount},
					},
					UpdateExpression:    aws.String("SET amount = amount + :amount, Version = :newVersion"),
					ConditionExpression: aws.String("attribute_exists(AccountID) AND TenantID = :tenantID"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
						":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
						":tenantID":   &types.AttributeValueMemberS{Value: trEntry.ToTenantID},
					},
				},
			},
			{Put: &types.Put{
				TableName: aws.String(LedgerTable),
				Item:      avCredit,
			}},
		},
	}

	_, err = dbSvc.TransactWriteItems(context, creditInput)
	if err != nil {
		rollbackInput := &dynamodb.UpdateItemInput{
			TableName: aws.String(NilUsers),
			Key: map[string]types.AttributeValue{
				"TenantID":  &types.AttributeValueMemberS{Value: trEntry.FromTenantID},
				"AccountID": &types.AttributeValueMemberS{Value: trEntry.FromAccount},
			},
			UpdateExpression:    aws.String("SET amount = amount + :amount, Version = :newVersion"),
			ConditionExpression: aws.String("attribute_not_exists(Version) OR Version = :oldVersion"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":amount":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", trEntry.Amount)},
				":oldVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(sender.Version, 10)},
				":newVersion": &types.AttributeValueMemberN{Value: strconv.FormatInt(getCurrentTimestamp(), 10)},
			},
		}

		_, rollbackErr := dbSvc.UpdateItem(context, rollbackInput)
		if rollbackErr != nil {
			panic(fmt.Errorf("failed to rollback debit for user %s: %v", trEntry.FromAccount, rollbackErr))
		}

		transactionStatus = 1
		if err := SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus); err != nil {
			panic(err)
		}
		response = NilResponse{
			Status:    "error",
			Code:      "credit_failed",
			Message:   fmt.Sprintf("Failed to credit to balance for user %s", trEntry.ToAccount),
			Details:   fmt.Sprintf("Error: %v", err),
			Timestamp: trEntry.Timestamp,
			Data: data{
				UUID:       trEntry.InitiatorUUID,
				SignedUUID: trEntry.SignedUUID,
			},
		}
		return response, fmt.Errorf("failed to credit to balance for user %s: %v", trEntry.ToAccount, err)
	}

	transactionStatus = 0
	if err := SaveToTransactionTable(dbSvc, combinedTenants, transaction, transactionStatus); err != nil {
		panic(err)
	}

	// now finally here: if cashout.provider was bok, then we should make a table for nil that will include:
	// - the transaction id
	// - the amount
	// - the currency
	// - the uuid
	// - the signed uuid
	// - the timestamp
	// - the actual to account
	// - the status of the transaction (pending, completed, failed), it will be first pending because we have not made the transaction yet, and then it will be completed when the transaction is completed
	// - the actual from account
	// - also if if if it was nil or empty string, we should also update the same data, so we can avail those data to our integrated partners to enquire about
	response = NilResponse{
		Status:  "success",
		Code:    "successful_transaction",
		Message: "Transaction initiated successfully.",
		Data: data{
			TransactionID: uid,
			Amount:        trEntry.Amount,
			Currency:      "SDG",
			UUID:          trEntry.InitiatorUUID,
			SignedUUID:    trEntry.SignedUUID,
		},
	}

	return response, nil
}

func GetEscrowTransactions(ctx context.Context, dbSvc *dynamodb.Client, tenantID string) ([]EscrowTransaction, error) {
	indexName := "FromTenantIDIndex"
	input := &dynamodb.QueryInput{
		TableName: aws.String("EscrowTransactions"),
		IndexName: aws.String(indexName), // Use the appropriate GSI
		KeyConditions: map[string]types.Condition{
			"FromTenantID": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: tenantID},
				},
			},
		},
	}

	result, err := dbSvc.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	var transactions []EscrowTransaction
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	log.Printf("the items are: %+v", transactions)
	return transactions, nil
}

func CreateServiceProvider(ctx context.Context, dbSvc *dynamodb.Client, serviceProvider ServiceProvider) error {
	// Marshal the ServiceProvider struct into a DynamoDB item
	if serviceProvider.Email == "" {
		return fmt.Errorf("email is required")
	}
	if serviceProvider.EscrowAccount == "" || serviceProvider.TenantID == "" {
		return fmt.Errorf("tenantID and escrowAccount are required")
	}
	if serviceProvider.Currency == "" {
		serviceProvider.Currency = "SDG"
	}

	serviceProvider.LastAccessed = time.Now().Format(time.RFC3339)
	item, err := attributevalue.MarshalMap(serviceProvider)
	if err != nil {
		return fmt.Errorf("failed to marshal service provider: %w", err)
	}

	// Create the PutItem input with a condition expression to ensure TenantID is unique
	input := &dynamodb.PutItemInput{
		TableName:           aws.String("ServiceProviders"),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(Email)"), // Ensure TenantID is unique
	}

	// Execute the PutItem operation
	_, err = dbSvc.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedErr *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedErr) {
			return fmt.Errorf("service provider with Email %s already exists", serviceProvider.TenantID)
		}
		return fmt.Errorf("failed to create service provider: %w", err)
	}

	return nil
}

func GetServiceProvider(ctx context.Context, dbSvc *dynamodb.Client, email string) (*ServiceProvider, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("ServiceProviders"),
		Key: map[string]types.AttributeValue{
			"Email": &types.AttributeValueMemberS{Value: email},
		},
	}

	result, err := dbSvc.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get service provider: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("service provider with Email %s not found", email)
	}

	var serviceProvider ServiceProvider
	if err := attributevalue.UnmarshalMap(result.Item, &serviceProvider); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service provider: %w", err)
	}

	return &serviceProvider, nil
}

func UpdateServiceProvider(ctx context.Context, dbSvc *dynamodb.Client, email string, svcProvider ServiceProvider) error {
	// Initialize an empty update expression and attribute values map
	updateExpression := "SET"
	expressionAttributeValues := make(map[string]types.AttributeValue)

	// Conditionally add to the update expression and attribute values
	if svcProvider.WebhookURL != "" {
		updateExpression += " WebhookURL = :webhook_url,"
		expressionAttributeValues[":webhook_url"] = &types.AttributeValueMemberS{Value: svcProvider.WebhookURL}
	}

	if svcProvider.TailscaleURL != "" {
		updateExpression += " TailscaleURL = :tailscale_url,"
		expressionAttributeValues[":tailscale_url"] = &types.AttributeValueMemberS{Value: svcProvider.TailscaleURL}
	}

	if svcProvider.PublicKey != "" {
		updateExpression += " PublicKey = :public_key,"
		expressionAttributeValues[":public_key"] = &types.AttributeValueMemberS{Value: svcProvider.PublicKey}
	}

	// Trim the trailing comma from the update expression
	updateExpression = updateExpression[:len(updateExpression)-1]

	// If there's nothing to update, return early
	if len(expressionAttributeValues) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Create the update item input with the dynamically built expression
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("ServiceProviders"),
		Key: map[string]types.AttributeValue{
			"Email": &types.AttributeValueMemberS{Value: email},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              types.ReturnValueUpdatedNew,
	}

	_, err := dbSvc.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update service provider: %w", err)
	}

	return nil
}

func ReverseEscrowTransferCredits(context context.Context, dbSvc *dynamodb.Client, es EscrowTransaction) error {
	// Create a new EscrowTransaction with reversed From and To accounts
	reversedEs := EscrowTransaction{
		FromAccount:   ESCROW_ACCOUNT,
		ToAccount:     es.FromAccount,
		Amount:        es.Amount,
		InitiatorUUID: es.InitiatorUUID,
		FromTenantID:  ESCROW_TENANT,
		ToTenantID:    es.FromTenantID,
	}

	// Call EscrowTransferCredits with the reversed transaction
	if _, err := EscrowTransferCredits(context, dbSvc, reversedEs); err != nil {
		return fmt.Errorf("failed to reverse escrow transfer: %w", err)
	}

	return nil
}

// StoreLocalWebhooks saves transactions in webhooks into a state so that it is retrievable later
func StoreLocalWebhooks(ctx context.Context, dbSvc *dynamodb.Client, serviceProvider string, transaction EscrowTransaction) error {
	item, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		// reverse the transfer here if fails!
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(ServiceProvidersTransactions), // save it in escrow transactions table
		Item:      item,
	}
	if _, err := dbSvc.PutItem(ctx, input); err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

func parseTimeInput(input string) (int64, error) {
	// If input is empty, return current time
	if input == "" {
		return time.Now().Unix(), nil
	}

	// Try parsing as Unix timestamp
	timestamp, err := strconv.ParseInt(input, 10, 64)
	if err == nil {
		return timestamp, nil
	}

	// If not a timestamp, try parsing as a date string
	t, err := time.Parse(time.RFC3339, input)
	if err == nil {
		return t.Unix(), nil
	}

	// If all else fails, return an error
	return 0, fmt.Errorf("unable to parse time input: %s", input)
}

func QueryServiceProviderTransactions(ctx context.Context, svc *dynamodb.Client, serviceProvider, startDateStr, endDateStr string, pageSize int32, lastEvaluatedKey map[string]types.AttributeValue) (*QueryResultEscrowWebhookTable, error) {
	startTimestamp, err := parseTimeInput(startDateStr)
	if err != nil {
		log.Printf("Warning: invalid start date (%s), using 1 month ago as default", startDateStr)
		startTimestamp = time.Now().AddDate(0, -1, 0).Unix()
	}

	endTimestamp, err := parseTimeInput(endDateStr)
	if err != nil {
		log.Printf("Warning: invalid end date (%s), using current time as default", endDateStr)
		endTimestamp = time.Now().Unix()
	}

	// Ensure startTimestamp is before endTimestamp
	if startTimestamp > endTimestamp {
		startTimestamp, endTimestamp = endTimestamp, startTimestamp
	}

	input := &dynamodb.QueryInput{
		TableName:              aws.String("ServiceProviderTransactions"),
		IndexName:              aws.String("ServiceProviderDateIndex"),
		KeyConditionExpression: aws.String("#sp = :sp AND #td BETWEEN :start AND :end"),
		ExpressionAttributeNames: map[string]string{
			"#sp": "ServiceProvider",
			"#td": "TransactionDate",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sp":    &types.AttributeValueMemberS{Value: serviceProvider},
			":start": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", startTimestamp)},
			":end":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", endTimestamp)},
		},
		Limit:             aws.Int32(pageSize),
		ExclusiveStartKey: lastEvaluatedKey,
	}

	result, err := svc.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %v", err)
	}

	var transactions []EscrowTransaction
	err = attributevalue.UnmarshalListOfMaps(result.Items, &transactions)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB result: %v", err)
	}

	return &QueryResultEscrowWebhookTable{
		Transactions:     transactions,
		LastEvaluatedKey: result.LastEvaluatedKey,
		HasMorePages:     len(result.LastEvaluatedKey) > 0,
	}, nil
}

func GetEscrowTransactionByUUID(ctx context.Context, svc *dynamodb.Client, uuid string) ([]EscrowTransaction, error) {
	// Prepare the query input
	input := &dynamodb.QueryInput{
		TableName: aws.String(EscrowTransactionsTable),

		KeyConditions: map[string]types.Condition{
			"UUID": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: uuid},
				},
			},
		},
	}

	// Execute the query
	result, err := svc.Query(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to query item, %v", err)
		return nil, err
	}

	// Parse the result
	var transactions []EscrowTransaction
	err = attributevalue.UnmarshalListOfMaps(result.Items, &transactions)
	if err != nil {
		log.Fatalf("failed to unmarshal query result, %v", err)
		return nil, err
	}

	return transactions, nil
}
