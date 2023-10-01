package ledger

import (
	"log"
	"reflect"
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
		{"test-credit", args{db: dbSvc, accountID: "249_ACCT_1", amount: 1000000}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RecordDebit(tt.args.db, tt.args.accountID, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("RecordDebit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitializeLedger(t *testing.T) {
	type args struct {
		accessKey string
		secretKey string
		region    string
	}
	tests := []struct {
		name    string
		args    args
		want    *dynamodb.DynamoDB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitializeLedger(tt.args.accessKey, tt.args.secretKey, tt.args.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitializeLedger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitializeLedger() = %v, want %v", got, tt.want)
			}
		})
	}
}
