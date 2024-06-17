package ledger

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/segmentio/ksuid"
)

func TestEscrowRequest(t *testing.T) {
	type args struct {
		context context.Context
		dbSvc   *dynamodb.Client
		esEntry EscrowEntry
	}
	tests := []struct {
		name    string
		args    args
		want    NilResponse
		wantErr bool
	}{
		// {"test escrow payment", args{context.TODO(), _dbSvc, EscrowEntry{FromAccount: "0111493885", ToAccount: ESCROW_ACCOUNT,
		// 	Amount: 4, ToTenantID: ESCROW_TENANT, FromTenantID: "nil", InitiatorUUID: ksuid.New().String()}},
		// 	NilResponse{}, false},
		{"test nonil-nil", args{context.TODO(), _dbSvc, EscrowEntry{FromAccount: "0111493885", ToAccount: "0965256869",
			Amount: 1, ToTenantID: "nil", FromTenantID: "nonil", InitiatorUUID: ksuid.New().String()}},
			NilResponse{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EscrowRequest(tt.args.context, tt.args.dbSvc, tt.args.esEntry)
			if (err != nil) != tt.wantErr {
				t.Errorf("EscrowRequest() error = %v, wantErr %v", err, tt.wantErr)

			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscrowRequest() = %v, want %v", got, tt.want)
			}
			if balance, err := InquireBalance(context.TODO(), _dbSvc, "nil", "0965256869"); err != nil || balance != 4 {
				t.Errorf("EscrowRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
