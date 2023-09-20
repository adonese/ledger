package ledger

import (
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var sess *session.Session
var err error
var dbSvc *dynamodb.DynamoDB

func TestRecordDebit(t *testing.T) {
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
	})
	if err != nil {
		log.Fatal("Failed to create DynamoDB session:", err)
	}

	// Create a DynamoDB client
	dbSvc := dynamodb.New(sess)
	type args struct {
		db        *dynamodb.DynamoDB
		accountID string
		amount    float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test-credit", args{db: dbSvc, accountID: "adonese", amount: 12}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RecordDebit(tt.args.db, tt.args.accountID, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("RecordDebit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
