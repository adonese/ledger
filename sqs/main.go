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

var webhookURL string

func init() {
	webhookURL = "https://dapi.nil.sd/webhook" // Set your webhook URL here
}

func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		var transaction ledger.EscrowTransaction
		err := json.Unmarshal([]byte(message.Body), &transaction)
		if err != nil {
			log.Printf("Failed to unmarshal SQS message: %v", err)
			continue
		}

		err = sendWebhookNotification(transaction)
		if err != nil {
			log.Printf("Failed to send webhook notification: %v", err)
			return err
		}
	}
	return nil
}

func sendWebhookNotification(transaction ledger.EscrowTransaction) error {
	transaction.Comment = "i'm coming via SQS message!"
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
	lambda.Start(handleRequest)
}
