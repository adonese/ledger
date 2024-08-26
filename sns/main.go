package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "embed"

	"github.com/adonese/ledger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var _dbSvc *dynamodb.Client

//go:embed priv.pem
var privKey []byte

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

// 2024-08-10T12:40:43Z app[90801932cd7018] ams [info]{"level":"info","method":"POST","path":"/webhook","status":200,"client_ip":"3.231.203.116","latency":0.13541,
// "request_body":{"transaction_id":"2kT3FZyVVUGn2j75pgzX4UA0lPL","from_account":"NIL_ESCROW_ACCOUNT","to_account":"0965256869","amount":4,"time":1723293636,"status":1,
// "from_tenant_id":"ESCROW_TENANT","to_tenant_id":"nil","uuid":"2kT3Fdzmy3LDyj9zP081dIXp7fQ","timestamp":"2024-08-10T12:40:38Z"},"response_body":,"time":"2024-08-10T12:40:43Z",
// "message":"request completed"}
func sendWebhookNotification(transaction ledger.EscrowTransaction) error {

	log.Printf("the transaction before newEscrowTransaction is: %+v", transaction)
	if err := ledger.StoreLocalWebhooks(context.TODO(), _dbSvc, transaction.ServiceProvider, transaction); err != nil {
		return fmt.Errorf("failed to store webhook: %w", err)
	}

	webhookURL := "https://dapi.nil.sd/webhook"
	log.Printf("the request as we've got it is: %+v", transaction)

	hookTransaction := NewEscrowTransactionWrapper(transaction)

	payload, err := json.Marshal(hookTransaction)
	if err != nil {
		return err
	}

	log.Printf("the transaction after conversion is: %+v", transaction)

	entry, err := ledger.GetServiceProvider(context.TODO(), _dbSvc, transaction.ServiceProvider)
	if err != nil {
		log.Printf("the error in lambda is: %v", err)
	}
	log.Printf("the entry is: %v", entry)
	if entry.WebhookURL != "" {
		webhookURL = entry.WebhookURL
	}

	nilSignature, err := sign(hookTransaction.InitiatorUUID, privKey)
	if err != nil {
		log.Printf("we failed to sign the transaction: %v", err)
	}

	// webhookURL = "https://dapi.nil.sd/webhook" //FIXME temporarily just to log the transaction

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", nilSignature)

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
	log.Println("i'm the actual sns")

	lambda.Start(handleSNSEvent)
}
