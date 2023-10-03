package ledger

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type Record struct {
	AccountID     string `json:"AccountID"`
	Amount        int    `json:"Amount"`
	OperationType string `json:"OperationType"`
}

// HandleDynamoDBStream handles sending notifications to nil users after payments
func HandleDynamoDBStream(ctx context.Context, event events.DynamoDBEvent) error {

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	sesSvc := ses.NewFromConfig(cfg)

	for _, record := range event.Records {
		newImage := record.Change.NewImage

		// Extract the necessary data from the new image
		accountID := newImage["AccountID"].String()
		amount := newImage["Amount"].String()
		opType := newImage["Message"].String()
		tranID := newImage["TransactionID"]

		op := "added to"
		if opType == "debit" {
			op = "deducted from"
		}

		message := fmt.Sprintf("The amount %s has been %s your account: %s\nTransaction ID: %s", amount, op, accountID, tranID)

		// Send email to the recipient
		err := SendEmail(sesSvc, Message{To: "mmbusif@gmail.com", Body: message, Subject: "Transaction Delivery"})
		if err != nil {
			log.Println("Error sending email:", err)
			return err
		}

		// // Update DynamoDB record to mark it as processed
		// _, err = dbSvc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		// 	TableName: aws.String("LedgerTable"),
		// 	Key: map[string]types.AttributeValue{
		// 		"Id": &types.AttributeValueMemberS{Value: newImage["Id"].String()},
		// 	},
		// 	ExpressionAttributeNames: map[string]string{
		// 		"#P": "Processed",
		// 	},
		// 	ExpressionAttributeValues: map[string]types.AttributeValue{
		// 		":p": &types.AttributeValueMemberBOOL{Value: true},
		// 	},
		// 	UpdateExpression: aws.String("SET #P = :p"),
		// })
		// if err != nil {
		// 	log.Println("Error updating DynamoDB record:", err)
		// 	return err
		// }
	}

	return nil
}

func SendSMS(sms SMS) error {
	log.Printf("the message is: %+v", sms)
	v := url.Values{}
	v.Add("api_key", sms.APIKey)
	v.Add("from", sms.Sender)
	v.Add("to", "249"+strings.TrimPrefix(sms.Mobile, "0"))
	v.Add("sms", sms.Message+"\n\n"+sms.Message)
	url := sms.Gateway + v.Encode()
	log.Printf("the url is: %+v", url)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("The error is: %+v", err)
		return err
	}
	log.Printf("The response body is: %v", res)
	return nil
}

func SendEmail(sesSvc *ses.Client, msg Message) error {

	// Specify the email details
	sender := "info.payment@nil.sd"
	recipient := "adonese@nil.sd"
	subject := msg.Subject
	htmlBody := msg.Body
	textBody := msg.Body

	// Create the email input
	emailInput := &ses.SendEmailInput{
		Destination: &sestypes.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &sestypes.Message{
			Body: &sestypes.Body{
				Html: &sestypes.Content{
					Data: aws.String(htmlBody),
				},
				Text: &sestypes.Content{
					Data: aws.String(textBody),
				},
			},
			Subject: &sestypes.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	// Send the email
	resp, err := sesSvc.SendEmail(context.TODO(), emailInput)
	if err != nil {
		log.Printf("failed to send email: %v", err)
		return err
	}
	// Print the message ID if the email was sent successfully
	fmt.Printf("Email sent! Message ID: %s\n", *resp.MessageId)
	return nil

}
