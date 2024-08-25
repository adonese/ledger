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
	log.Println("the SQSProcessor is launched")

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
	// in the initial request we will have:
	// - from account
	// - to account
	// - from tenat to tenat
	log.Println("Hello, World! SQS is being loaded!")
	_, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	for _, record := range event.Records {
		if record.EventName == "INSERT" {
			dynamoRecord := record.Change.NewImage

			log.Printf("the record is: %+v", dynamoRecord)

			// Convert events.DynamoDBAttributeValue to dynamodb.AttributeValue
			convertedRecord := make(map[string]types.AttributeValue)
			for k, v := range dynamoRecord {
				convertedRecord[k] = ledger.ConvertToSDKAttributeValue(v)
			}

			var transaction ledger.EscrowTransaction
			err := attributevalue.UnmarshalMap(convertedRecord, &transaction)
			if err != nil {
				log.Printf("failed to unmarshal DynamoDB record, %v", err)
				continue
			}
			log.Printf("the transaction entry is: %+v", transaction)
			if transaction.Status == ledger.StatusCompleted || transaction.Status == ledger.StatusFailed { // both are final states
				continue
			}
			if transaction.CashoutProvider == "nil" { // this is guaranteed to be populated

				esTransaction := ledger.EscrowTransaction{
					FromAccount:         transaction.TransientAccount,
					FromTenantID:        transaction.TransientTenant,
					ToAccount:           transaction.ToAccount,
					ToTenantID:          transaction.ToTenantID,
					Amount:              transaction.Amount,
					Comment:             "Cashout",
					InitiatorUUID:       transaction.InitiatorUUID,
					Timestamp:           transaction.Timestamp,
					SystemTransactionID: transaction.SystemTransactionID,
					CashoutProvider:     transaction.CashoutProvider,
					TransientAccount:    transaction.TransientAccount,
					TransientTenant:     transaction.TransientTenant,
					ServiceProvider:     transaction.ServiceProvider,
					TransactionDate:     transaction.TransactionDate,
				}

				esTransaction.Status = ledger.StatusCompleted
				if _, err := ledger.EscrowTransferCredits(context.TODO(), _dbSvc, esTransaction); err != nil {
					log.Printf("the error in sending transaction is: %v", err)
					esTransaction.Status = ledger.StatusFailed
					// You should reverse operation here!
					reversedTrans := ledger.EscrowTransaction{
						FromAccount:         transaction.TransientAccount,
						FromTenantID:        transaction.TransientTenant,
						ToAccount:           transaction.FromAccount,
						ToTenantID:          transaction.FromTenantID,
						Amount:              transaction.Amount,
						InitiatorUUID:       transaction.InitiatorUUID,
						SystemTransactionID: transaction.SystemTransactionID,
					}
					log.Printf("the reverse request is: %+v", reversedTrans)
					if _, err := ledger.EscrowTransferCredits(context.TODO(), _dbSvc, reversedTrans); err != nil {
						log.Printf("WE should fix this: %v", err)
					}

				}
				log.Printf("the request we're sending to sns is: %+v", esTransaction)
				// In handleRequest function, after updating the transaction

				if err := publishToSNSTopic(esTransaction); err != nil {
					log.Printf("failed to publish to SNS topic: %v", err)
				}

				// Now, i want to amend that table again to make the status as completed.
				updateItem(context.TODO(), _dbSvc, esTransaction)
			} else {
				log.Println("im unable to hit the cashout provider")
				log.Printf("log for saving transaction is: %+v", ledger.StoreLocalWebhooks(context.TODO(), _dbSvc, transaction.ServiceProvider, transaction))
			}

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

func main() {
	lambda.Start(handleRequest)
}
