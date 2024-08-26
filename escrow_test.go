package ledger

import (
	"context"
	"reflect"
	"testing"
	"time"

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
			ServiceProvider: "oss@pynil.com",
			Amount:          1, ToTenantID: "nil", FromTenantID: "nonil", InitiatorUUID: "fff", PaymentReference: "1234567890"},
		},
			NilResponse{}, false},
		{"test nonil-nil", args{context.TODO(), _dbSvc, EscrowEntry{
			CashoutProvider: "bok",
			FromAccount:     "0111493885", ToAccount: "0965256869",
			ServiceProvider: "oss@pynil.com",
			Amount:          2, ToTenantID: "nil", FromTenantID: "nonil", InitiatorUUID: "fff"},
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
		{"test nil tenant", args{context.TODO(), _dbSvc, "oss@pynil.com"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetServiceProvider(tt.args.ctx, tt.args.dbSvc, tt.args.tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServiceProvider() = %+v, want %v", got, tt.want)
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
		{"test nil tenant", args{context.TODO(), _dbSvc, "oss11@pynil.com", ServiceProvider{WebhookSigningKey: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6n9XrRSSZZM46mmsE3F0qVjnFgcGKySy+jaTuOX2QjNI8qysbyL/hoDqhYhmOoPPbwn18JO2Ochw+EXcbKnR9qAPIu8CEeUweo0LG+Cv5SL/WBI2kaWpDz3fMSzw+Hanf6hRqm7jsWR/RV5qPI73IdBJ3gfdUpv9Ta8uzk7HOwIuR30Ja7pLKleIf5HFt56uFx8dxAofv7I8cc0NFbhKa7A937/DyqQG7vE+CGlF2MZPdMw0HMfOCxFWGekVwlrwkmdxjgtaNYJrtxHmzHOwVcnT7/7kGZrZ5GxefuV6eMo2ed4y0/QF/wzyZuBCQATkL962xiELcGkjzIIbcb1YlQIDAQAB"}}, false},
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
		{"test nil tenant", args{context.TODO(), _dbSvc, ServiceProvider{TenantID: "11nil", WebhookURL: "http://localhost:8080", Email: "oss@pynil.com"}}, false},
		{"test nil tenant", args{context.TODO(), _dbSvc, ServiceProvider{TenantID: "nil", WebhookURL: "http://localhost:8089", Email: "oss@pynil.com"}}, false},
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
		"PaymentReference": events.NewStringAttribute("1234567890"),
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
	assert.Equal(t, "1234567890", transaction.PaymentReference)
}

func TestQueryServiceProviderTransactions(t *testing.T) {
	type args struct {
		ctx              context.Context
		svc              *dynamodb.Client
		serviceProvider  string
		startDate        time.Time
		endDate          time.Time
		pageSize         int32
		lastEvaluatedKey map[string]types.AttributeValue
	}
	tests := []struct {
		name    string
		args    args
		want    *QueryResultEscrowWebhookTable
		wantErr bool
	}{
		{"test nil tenant", args{context.TODO(), _dbSvc, "oss@pynil.com", time.Now().Add(-24 * 30 * 10 * time.Hour), time.Now(), 100, nil}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryServiceProviderTransactions(tt.args.ctx, tt.args.svc, tt.args.serviceProvider, tt.args.startDate.Format(time.RFC3339), tt.args.endDate.Format(time.RFC3339), tt.args.pageSize, tt.args.lastEvaluatedKey)
			if err != nil {
				t.Errorf("QueryServiceProviderTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QueryServiceProviderTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEscrowTransactionByUUID(t *testing.T) {
	type args struct {
		ctx  context.Context
		svc  *dynamodb.Client
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    *EscrowTransaction
		wantErr bool
	}{
		// get escrow transaction by uuid
		{"test nil tenant", args{context.TODO(), _dbSvc, "fc1486cd-b245-4f34-81e7-c87c784a40f5"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEscrowTransactionByUUID(tt.args.ctx, tt.args.svc, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEscrowTransactionByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEscrowTransactionByUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDuplicateEscrowTransaction(t *testing.T) {
	type args struct {
		ctx  context.Context
		svc  *dynamodb.Client
		uuid string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test nil tenant", args{context.TODO(), _dbSvc, "fc1486cd-b245-4f34-81e7-c87c784a40f5"}, true},
		{"test nil tenant", args{context.TODO(), _dbSvc, "fff"}, true},
		{"test nil tenant", args{context.TODO(), _dbSvc, "fff333"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDuplicateEscrowTransaction(tt.args.ctx, tt.args.svc, tt.args.uuid); got != tt.want {
				t.Errorf("IsDuplicateEscrowTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}
