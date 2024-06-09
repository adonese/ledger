package ledger

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/segmentio/ksuid"
)

type QRPaymentRequest struct {
	TenantID     string  `json:"TenantID"`
	PaymentID    string  `json:"PaymentID"`
	AccountID    string  `json:"AccountID"`
	Amount       float64 `json:"Amount"`
	Status       string  `json:"Status"`
	UUID         string  `json:"UUID"`
	CreationDate int64   `json:"CreationDate"`
	FromAccount  string  `json:"from_account"`
	ToAccount    string  `json:"to_account"`
}

func (qr *QRPaymentRequest) IsPaid() bool {
	return qr.Status == "COMPLETED"
}

func GenerateQRPayment(ctx context.Context, dbSvc *dynamodb.Client, tenantID, accountID string, amount float64) (*QRPaymentRequest, error) {
	paymentID := ksuid.New().String()
	uuid := ksuid.New().String()
	timestamp := time.Now().UTC().Unix()

	qrPayment := QRPaymentRequest{
		TenantID:     tenantID,
		PaymentID:    paymentID,
		AccountID:    accountID,
		Amount:       amount,
		Status:       "PENDING",
		UUID:         uuid,
		CreationDate: timestamp,
		ToAccount:    accountID,
	}

	av, err := attributevalue.MarshalMap(qrPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal QR payment request: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("QRPaymentsTable"),
		Item:      av,
	}

	_, err = dbSvc.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR payment request: %v", err)
	}

	return &qrPayment, nil
}

func InquireQRPayment(ctx context.Context, dbSvc *dynamodb.Client, tenantID, paymentID string) (*QRPaymentRequest, error) {
	key := map[string]types.AttributeValue{
		"TenantID":  &types.AttributeValueMemberS{Value: tenantID},
		"PaymentID": &types.AttributeValueMemberS{Value: paymentID},
	}

	result, err := dbSvc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("QRPaymentsTable"),
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to inquire QR payment: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("QR payment %s does not exist", paymentID)
	}

	var qrPayment QRPaymentRequest
	err = attributevalue.UnmarshalMap(result.Item, &qrPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal QR payment: %v", err)
	}

	return &qrPayment, nil
}

func PerformQRPayment(ctx context.Context, dbSvc *dynamodb.Client, tenantID, paymentID, personPayingAccount string) error {
	qrPayment, err := InquireQRPayment(ctx, dbSvc, tenantID, paymentID)
	if err != nil {
		return err
	}

	if qrPayment.Status != "PENDING" {
		return fmt.Errorf("QR payment %s is not in PENDING status", paymentID)
	}

	trEntry := TransactionEntry{
		TenantID:      tenantID,
		FromAccount:   personPayingAccount,
		AccountID:     qrPayment.ToAccount,
		ToAccount:     qrPayment.AccountID,
		Amount:        qrPayment.Amount,
		InitiatorUUID: ksuid.New().String(),
	}

	response, err := TransferCredits(ctx, dbSvc, trEntry)
	if err != nil {
		return fmt.Errorf("failed to perform QR payment: %v", err)
	}

	log.Printf("the result of transfer is: %+v", response)

	updateExpression := "SET #st = :status"
	expressionAttributeNames := map[string]string{
		"#st": "Status",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: "COMPLETED"},
	}

	_, err = dbSvc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("QRPaymentsTable"),
		Key: map[string]types.AttributeValue{
			"TenantID":  &types.AttributeValueMemberS{Value: tenantID},
			"PaymentID": &types.AttributeValueMemberS{Value: paymentID},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	if err != nil {
		return fmt.Errorf("failed to update QR payment status: %v", err)
	}

	return nil
}

func GetAllQRPaymentsForUser(ctx context.Context, dbSvc *dynamodb.Client, tenantID, creatorAccountID string) ([]QRPaymentRequest, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("QRPaymentsTable"),
		IndexName:              aws.String("CreatorAccountIDIndex"),
		KeyConditionExpression: aws.String("TenantID = :tenantID AND CreatorAccountID = :creatorAccountID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tenantID":         &types.AttributeValueMemberS{Value: tenantID},
			":creatorAccountID": &types.AttributeValueMemberS{Value: creatorAccountID},
		},
	}

	result, err := dbSvc.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query QR payments: %v", err)
	}

	var qrPayments []QRPaymentRequest
	err = attributevalue.UnmarshalListOfMaps(result.Items, &qrPayments)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal QR payments: %v", err)
	}

	return qrPayments, nil
}
