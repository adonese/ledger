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
)

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

func sendWebhookNotification(transaction ledger.EscrowTransaction) error {
	webhookURL := "https://dapi.nil.sd/webhook"
	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
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

func main() {
	log.Println("The sns is launched again")
	lambda.Start(handleSNSEvent)
}
