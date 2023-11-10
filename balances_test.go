package ledger

import (
	"context"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

var _dbSvc *dynamodb.Client
var _sesSvc *ses.Client

func init() {
	var err error

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-north-1"),
	)
	if err != nil {
		log.Fatal("Failed to create DynamoDB session:", err)
	}

	_dbSvc = dynamodb.NewFromConfig(cfg)
	_sesSvc = ses.NewFromConfig(cfg)
}

func Test_transferCredits(t *testing.T) {

	type args struct {
		dbSvc         *dynamodb.Client
		fromAccountID string
		toAccountID   string
		amount        float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "adonese", dbSvc: _dbSvc, amount: 69}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TransferCredits(tt.args.dbSvc, tt.args.fromAccountID, tt.args.toAccountID, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("transferCredits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_inquireBalance(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		AccountID string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "249_ACCT_1"}, 30, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InquireBalance(tt.args.dbSvc, tt.args.AccountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("inquireBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("inquireBalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createAccountWithBalance(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		accountId string
		amount    float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "249_ACCT_1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccountWithBalance(tt.args.dbSvc, tt.args.accountId, (tt.args.amount)); (err != nil) != tt.wantErr {
				t.Errorf("createAccountWithBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckUser(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		accountId string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "adonese"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notF, err := CheckUsersExist(tt.args.dbSvc, []string{tt.args.accountId, "44322"})
			if err != nil {
				t.Errorf("there's an error: %v - notfound: %v", err, notF)
				return
			}
		})
	}
}
