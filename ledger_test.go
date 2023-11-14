package ledger

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var _AWS_REGION = "us-east-1"

func TestRecordDebit(t *testing.T) {

	type args struct {
		db        *dynamodb.Client
		accountID string
		amount    float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test-credit", args{db: _dbSvc, accountID: "249_ACCT_1", amount: 101}, false},
		{"test-credit", args{db: _dbSvc, accountID: "249_ACCT_1", amount: 101}, false},
		{"test-credit", args{db: _dbSvc, accountID: "249_ACCT_1", amount: 101}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RecordDebit(tt.args.db, tt.args.accountID, tt.args.amount); err != nil {
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
		name string
		args args
	}{
		{"test_initializing aws", args{accessKey: "", secretKey: "", region: _AWS_REGION}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitializeLedger(tt.args.accessKey, tt.args.secretKey, tt.args.region)
			if err != nil {
				t.Errorf("InitializeLedger() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("InitializeLedger() error = ledger is nil")
				return
			}
		})
	}
}
