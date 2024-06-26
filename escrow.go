package ledger

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/segmentio/ksuid"
)

// Status represents the status of a transaction
type Status int

// Define possible statuses
const (
	StatusPending Status = iota
	StatusCompleted
	StatusFailed
	StatusInProgress
)

const EscrowTransactionsTable = "EscrowTransactions"
const ESCROW_ACCOUNT = "NIL_ESCROW_ACCOUNT"
const ESCROW_TENANT = "ESCROW_TENANT"

// String returns the string representation of the Status
func (s Status) String() string {
	return [...]string{"Pending", "Completed", "Failed", "In Progress"}[s]
}

func EscrowRequest(context context.Context, dbSvc *dynamodb.Client, esEntry EscrowEntry) (NilResponse, error) {
	var response NilResponse

	timestamp := getCurrentTimestamp()
	transactionStatus := StatusPending

	uid := ksuid.New().String()

	es := EscrowTransaction{
		FromAccount:   esEntry.FromAccount,
		ToAccount:     ESCROW_ACCOUNT, // We need to make sure that ESCROW_ACCOUNT is an exception
		Amount:        esEntry.Amount,
		InitiatorUUID: esEntry.InitiatorUUID,
		FromTenantID:  esEntry.FromTenantID,
		ToTenantID:    ESCROW_TENANT,
	}

	if _, err := EscrowTransferCredits(context, dbSvc, es); err != nil {
		return NilResponse{}, err
	}

	es.Status = transactionStatus
	es.SystemTransactionID = uid
	es.TransactionDate = timestamp

	// // HERE we are supposed to ensure that from and to are actually matches what we want
	// Now, after you have done that, you should write these to Table
	// What we need here is a fully fledged EscrowTransaction
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
	}

	item, err := attributevalue.MarshalMap(esTransaction)
	if err != nil {
		return NilResponse{}, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(EscrowTransactionsTable),
		Item:      item,
	}

	if _, err := dbSvc.PutItem(context, input); err != nil {
		spew.Dump(item)
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

	// Fetch sender account
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

	debitEntry := LedgerEntry{
		TenantID:            trEntry.FromTenantID,
		AccountID:           trEntry.FromAccount,
		Amount:              trEntry.Amount,
		SystemTransactionID: uid,
		Type:                "debit",
		Time:                timestamp,
		InitiatorUUID:       trEntry.InitiatorUUID,
	}
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
