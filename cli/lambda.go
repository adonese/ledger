package main

import (
	"context"
	"encoding/json"
	"log"

	_ "embed"

	"github.com/adonese/ledger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

//go:embed .secrets.json
var secrets []byte

type Data struct {
	AwsKey    string `json:"AWS_ACCESS_KEY_ID"`
	AwsSecret string `json:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion string `json:"AWS_REGION"`
}

var data Data

var _dbSvc *dynamodb.Client
var _snsSvc *sns.Client

func init() {
	// Initialize SNS client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	_snsSvc = sns.NewFromConfig(cfg)
}

func publishToSNSTopic(transaction ledger.EscrowTransaction) error {
	message, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	input := &sns.PublishInput{
		Message:  aws.String(string(message)),
		TopicArn: aws.String(ledger.SNS_TOPIC),
	}

	_, err = _snsSvc.Publish(context.TODO(), input)
	return err
}

func init() {

	json.Unmarshal(secrets, &data)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(data.AwsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			data.AwsKey,
			data.AwsSecret,
			"",
		)),
	)
	if err != nil {
		log.Fatal("Failed to create DynamoDB session:", err)
	}

	_dbSvc = dynamodb.NewFromConfig(cfg)
}

func handleRequest(ctx context.Context, event events.DynamoDBEvent) {
	_, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	for _, record := range event.Records {
		if record.EventName == "INSERT" {
			dynamoRecord := record.Change.NewImage

			// Convert events.DynamoDBAttributeValue to dynamodb.AttributeValue
			convertedRecord := make(map[string]types.AttributeValue)
			for k, v := range dynamoRecord {
				convertedRecord[k] = convertToSDKAttributeValue(v)
			}

			var transaction ledger.EscrowTransaction
			err := attributevalue.UnmarshalMap(convertedRecord, &transaction)
			if err != nil {
				log.Printf("failed to unmarshal DynamoDB record, %v", err)
				continue
			}
			log.Printf("the transaction entry is: %+v", transaction)
			if transaction.Status == ledger.StatusCompleted {
				continue
			}

			var reversedTenant, reveresedAccount string // we use those to reverse the transaction
			reversedTenant = transaction.FromTenantID
			reveresedAccount = transaction.FromAccount

			// This is what we get
			// {SystemTransactionID:2hvsbAMI36eCWn1awIOQ6B9ExfP FromAccount:0111493885 ToAccount:NIL_ESCROW_ACCOUNT Amount:4 Comment: TransactionDate:1718485953
			//  Status:Pending FromTenantID:nonil ToTenantID:ESCROW_TENANT InitiatorUUID:2hvsb5L4EPLyEfqtrmC8kEfwtJq Timestamp: SignedUUID:}

			transaction.FromTenantID = ledger.ESCROW_TENANT
			transaction.FromAccount = ledger.ESCROW_ACCOUNT
			transaction.Status = ledger.StatusCompleted
			if _, err := ledger.EscrowTransferCredits(context.TODO(), _dbSvc, transaction); err != nil {
				log.Printf("the error in sending transaction is: %v", err)
				transaction.Status = ledger.StatusFailed
				// You should reverse operation here!
				tr := ledger.EscrowTransaction{
					FromAccount:   ledger.ESCROW_ACCOUNT,
					ToAccount:     reveresedAccount,
					FromTenantID:  ledger.ESCROW_TENANT,
					ToTenantID:    reversedTenant,
					Amount:        transaction.Amount,
					InitiatorUUID: transaction.InitiatorUUID,
				}
				log.Printf("the reverse request is: %+v", tr)
				if _, err := ledger.EscrowTransferCredits(context.TODO(), _dbSvc, tr); err != nil {
					log.Printf("WE should fix this: %v", err)
				}

			}
			// In handleRequest function, after updating the transaction
			if err := publishToSNSTopic(transaction); err != nil {
				log.Printf("failed to publish to SNS topic: %v", err)
			}

			// Now, i want to amend that table again to make the status as completed.
			updateItem(context.TODO(), _dbSvc, transaction)

		}
	}
}

func updateItem(ctx context.Context, dbSvc *dynamodb.Client, transaction ledger.EscrowTransaction) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(ledger.EscrowTransactionsTable),
		Key: map[string]types.AttributeValue{
			"UUID":          &types.AttributeValueMemberS{Value: transaction.InitiatorUUID},
			"TransactionID": &types.AttributeValueMemberS{Value: transaction.SystemTransactionID},
		},
		UpdateExpression: aws.String("SET #ts = :status"),
		ExpressionAttributeNames: map[string]string{
			"#ts": "TransactionStatus",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: transaction.Status.String()},
		},
	}

	_, err := dbSvc.UpdateItem(ctx, input)
	if err != nil {
		log.Printf("failed to update item in DynamoDB: %v", err)
	}
}

func convertToSDKAttributeValue(av events.DynamoDBAttributeValue) types.AttributeValue {
	switch av.DataType() {
	case events.DataTypeString:
		return &types.AttributeValueMemberS{Value: av.String()}
	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{Value: av.Number()}
	case events.DataTypeBinary:
		return &types.AttributeValueMemberB{Value: av.Binary()}
	case events.DataTypeStringSet:
		return &types.AttributeValueMemberSS{Value: av.StringSet()}
	case events.DataTypeNumberSet:
		return &types.AttributeValueMemberNS{Value: av.NumberSet()}
	case events.DataTypeBinarySet:
		return &types.AttributeValueMemberBS{Value: av.BinarySet()}
	case events.DataTypeMap:
		m := av.Map()
		mapAv := make(map[string]types.AttributeValue, len(m))
		for k, v := range m {
			mapAv[k] = convertToSDKAttributeValue(v)
		}
		return &types.AttributeValueMemberM{Value: mapAv}
	case events.DataTypeList:
		l := av.List()
		listAv := make([]types.AttributeValue, len(l))
		for i, v := range l {
			listAv[i] = convertToSDKAttributeValue(v)
		}
		return &types.AttributeValueMemberL{Value: listAv}
	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{Value: av.Boolean()}
	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{Value: av.IsNull()}
	default:
		return nil
	}
}

func main() {
	lambda.Start(handleRequest)
}
