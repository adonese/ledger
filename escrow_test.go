package ledger

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
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
		{"test nonil-nil", args{context.TODO(), _dbSvc, EscrowEntry{
			CashoutProvider: "nil",
			FromAccount:     "0111493885", ToAccount: "0965256869",
			Amount: 3, ToTenantID: "nil", FromTenantID: "nonil", InitiatorUUID: "this is my static uuid YOLO!"},
		},
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
				t.Errorf("GetEscrowTransactions() error = %+v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEscrowTransactions() = %+v, want %v", got, tt.want)
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

func TestUnmarshalDynamoDBEvent(t *testing.T) {
	// Create a sample DynamoDB event NewImage
	sampleNewImage := map[string]events.DynamoDBAttributeValue{
		"TransactionID":     events.NewStringAttribute("2l6scKOWFYa2BsXtGATe353uvEU"),
		"FromAccount":       events.NewStringAttribute("0111493885"),
		"ToAccount":         events.NewStringAttribute("0965256869"),
		"Amount":            events.NewNumberAttribute("5"),
		"Comment":           events.NewStringAttribute(""),
		"TransactionDate":   events.NewNumberAttribute("1724511938"),
		"TransactionStatus": events.NewStringAttribute(StatusPending.String()),
		"FromTenantID":      events.NewStringAttribute("nonil"),
		"ToTenantID":        events.NewStringAttribute("nil"),
		"UUID":              events.NewStringAttribute("this is my static uuid YOLO!"),
		"timestamp":         events.NewStringAttribute("2024-08-24T15:05:39Z"),
		"signed_uuid":       events.NewStringAttribute(""),
		"CashoutProvider":   events.NewNullAttribute(),
		"Beneficiary": events.NewMapAttribute(map[string]events.DynamoDBAttributeValue{
			"AccountID": events.NewStringAttribute(""),
			"FullName":  events.NewStringAttribute(""),
			"Mobile":    events.NewStringAttribute(""),
			"Provider":  events.NewStringAttribute(""),
			"Address":   events.NewStringAttribute(""),
		}),
		"TransientAccount": events.NewStringAttribute("NIL_ESCROW_ACCOUNT"),
		"TransientTenant":  events.NewStringAttribute("ESCROW_TENANT"),
	}

	// Convert events.DynamoDBAttributeValue to types.AttributeValue
	convertedRecord := make(map[string]types.AttributeValue)
	for k, v := range sampleNewImage {
		convertedRecord[k] = ConvertToSDKAttributeValue(v)
	}

	// Unmarshal the converted record
	var transaction EscrowTransaction
	err := attributevalue.UnmarshalMap(convertedRecord, &transaction)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "2l6scKOWFYa2BsXtGATe353uvEU", transaction.SystemTransactionID)
	assert.Equal(t, "0111493885", transaction.FromAccount)
	assert.Equal(t, "0965256869", transaction.ToAccount)
	assert.Equal(t, float64(5), transaction.Amount)
	assert.Equal(t, "", transaction.Comment)
	assert.Equal(t, int64(1724511938), transaction.TransactionDate)
	assert.Equal(t, "Pending", transaction.Status.String())
	assert.Equal(t, "nonil", transaction.FromTenantID)
	assert.Equal(t, "nil", transaction.ToTenantID)
	assert.Equal(t, "this is my static uuid YOLO!", transaction.InitiatorUUID)
	assert.Equal(t, "2024-08-24T15:05:39Z", transaction.Timestamp)
	assert.Equal(t, "", transaction.SignedUUID)
	assert.Equal(t, "", transaction.CashoutProvider)
	assert.Equal(t, Beneficiary{}, transaction.Beneficiary)
	assert.Equal(t, "NIL_ESCROW_ACCOUNT", transaction.TransientAccount)
	assert.Equal(t, "ESCROW_TENANT", transaction.TransientTenant)
}
