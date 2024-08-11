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
			Amount: 4, ToTenantID: "nil", FromTenantID: "nonil", InitiatorUUID: ksuid.New().String()}},
			NilResponse{}, false},
		{"test nonil-nil", args{context.TODO(), _dbSvc, EscrowEntry{ToAccount: "0111493885", FromAccount: "0965256869",
			Amount: 4, ToTenantID: "nonil", FromTenantID: "nil", InitiatorUUID: ksuid.New().String()}},
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

func TestGetEscrowTransactions(t *testing.T) {
	type args struct {
		ctx      context.Context
		dbSvc    *dynamodb.Client
		tenantID string
	}
	tests := []struct {
		name    string
		args    args
		want    []EscrowTransaction
		wantErr bool
	}{
		{"test nil tenant", args{context.TODO(), _dbSvc, "nil"}, []EscrowTransaction{}, false},
		{"test nonil tenant", args{context.TODO(), _dbSvc, "nonil"}, []EscrowTransaction{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEscrowTransactions(tt.args.ctx, tt.args.dbSvc, tt.args.tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEscrowTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEscrowTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServiceProvider(t *testing.T) {
	type args struct {
		ctx      context.Context
		dbSvc    *dynamodb.Client
		tenantID string
	}
	tests := []struct {
		name    string
		args    args
		want    *ServiceProvider
		wantErr bool
	}{
		{"test nil tenant", args{context.TODO(), _dbSvc, "nil"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetServiceProvider(tt.args.ctx, tt.args.dbSvc, tt.args.tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServiceProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateServiceProvider(t *testing.T) {
	type args struct {
		ctx         context.Context
		dbSvc       *dynamodb.Client
		tenantID    string
		svcProvider ServiceProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test nil tenant", args{context.TODO(), _dbSvc, "nil", ServiceProvider{WebhookURL: "http://localhost:8080"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateServiceProvider(tt.args.ctx, tt.args.dbSvc, tt.args.tenantID, tt.args.svcProvider); (err != nil) != tt.wantErr {
				t.Errorf("UpdateServiceProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateServiceProvider(t *testing.T) {
	type args struct {
		ctx             context.Context
		dbSvc           *dynamodb.Client
		serviceProvider ServiceProvider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// create with public key
		{"test nil tenant", args{context.TODO(), _dbSvc, ServiceProvider{TenantID: "11nil", WebhookURL: "http://localhost:8080"}}, false},
		{"test nil tenant", args{context.TODO(), _dbSvc, ServiceProvider{TenantID: "nil", WebhookURL: "http://localhost:8089"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateServiceProvider(tt.args.ctx, tt.args.dbSvc, tt.args.serviceProvider); (err != nil) != tt.wantErr {
				t.Errorf("CreateServiceProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
