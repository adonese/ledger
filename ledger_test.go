package ledger

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

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
		{"test-credit", args{db: _dbSvc, accountID: "249_ACCT_1", amount: 1000000}, false},
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
		want    *dynamodb.Client
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
