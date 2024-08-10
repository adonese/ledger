package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adonese/ledger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var _dbSvc *dynamodb.Client

func handleSNSEvent(ctx context.Context, snsEvent events.SNSEvent) {
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		log.Printf("the record is: %+v", snsRecord)
		var transaction ledger.EscrowTransaction
		err := json.Unmarshal([]byte(snsRecord.Message), &transaction)
		if err != nil {
			log.Printf("failed to unmarshal SNS message: %v", err)
			continue
		}

		err = sendWebhookNotification(transaction)
		if err != nil {
			log.Printf("failed to send webhook notification: %v", err)
			// Handle error as needed, potentially adding retry logic
		}
	}
}

func getServiceProvider(ctx context.Context, client *dynamodb.Client, tenantID string) (*ServiceProvider, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("ServiceProviders"),
		Key: map[string]types.AttributeValue{
			"TenantID": &types.AttributeValueMemberS{Value: tenantID},
		},
	}

	result, err := client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("no item found for tenant_id: %s", tenantID)
	}

	serviceProvider := &ledger.ServiceProvider{
		TenantID:     tenantID,
		WebhookURL:   result.Item["WebhookURL"].(*types.AttributeValueMemberS).Value,
		TailscaleURL: result.Item["TailscaleURL"].(*types.AttributeValueMemberS).Value,
		LastAccessed: result.Item["LastAccessed"].(*types.AttributeValueMemberS).Value,
	}

	return serviceProvider, nil
}

// 2024-08-10T12:40:43Z app[90801932cd7018] ams [info]{"level":"info","method":"POST","path":"/webhook","status":200,"client_ip":"3.231.203.116","latency":0.13541,
// "request_body":{"transaction_id":"2kT3FZyVVUGn2j75pgzX4UA0lPL","from_account":"NIL_ESCROW_ACCOUNT","to_account":"0965256869","amount":4,"time":1723293636,"status":1,
// "from_tenant_id":"ESCROW_TENANT","to_tenant_id":"nil","uuid":"2kT3Fdzmy3LDyj9zP081dIXp7fQ","timestamp":"2024-08-10T12:40:38Z"},"response_body":,"time":"2024-08-10T12:40:43Z",
// "message":"request completed"}
func sendWebhookNotification(transaction ledger.EscrowTransaction) error {

	webhookURL := "https://dapi.nil.sd/webhook"

	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	entry, err := getServiceProvider(context.TODO(), _dbSvc, transaction.FromTenantID)
	if err != nil {
		log.Printf("the error in lambda is: %v", err)
	}
	log.Printf("the entry is: %v", entry)
	if entry.WebhookURL != "" {
		webhookURL = entry.WebhookURL
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	return nil
}

func init() {
	log.Println("The sns is launched")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create a DynamoDB client
	client := dynamodb.NewFromConfig(cfg)
	_dbSvc = client
}

func main() {
	log.Println("The sns is launched again")

	lambda.Start(handleSNSEvent)
}
