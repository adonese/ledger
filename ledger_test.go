package ledger

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var _AWS_REGION = "us-east-1"

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

func TestDeleteAccount(t *testing.T) {
	type args struct {
		ctx       context.Context
		dbSvc     *dynamodb.Client
		tenantId  string
		accountId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test_delete_account", args{ctx: context.TODO(), dbSvc: _dbSvc, tenantId: "nil", accountId: "+1234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteAccount(tt.args.ctx, tt.args.dbSvc, tt.args.tenantId, tt.args.accountId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
