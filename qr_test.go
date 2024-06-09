package ledger

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

func TestQRPaymentFunctionalIntegration(t *testing.T) {

	tenantID := "nil"
	accountID := "0111493885"
	fromAccountID := "0111493888"
	amount := 100.0

	ctx := context.Background()

	// Step 1: Generate QR Payment
	qrPayment, err := GenerateQRPayment(ctx, _dbSvc, tenantID, accountID, amount)
	assert.NoError(t, err)
	assert.NotNil(t, qrPayment)

	// Step 2: Inquire about the generated QR Payment
	inquiredPayment, err := InquireQRPayment(ctx, _dbSvc, tenantID, qrPayment.PaymentID)
	assert.NoError(t, err)
	assert.NotNil(t, inquiredPayment)
	assert.Equal(t, qrPayment.TenantID, inquiredPayment.TenantID)
	assert.Equal(t, qrPayment.PaymentID, inquiredPayment.PaymentID)
	assert.Equal(t, qrPayment.AccountID, inquiredPayment.AccountID)
	assert.Equal(t, qrPayment.Amount, inquiredPayment.Amount)
	assert.Equal(t, "PENDING", inquiredPayment.Status)

	// Step 3: Perform QR Payment
	err = PerformQRPayment(ctx, _dbSvc, tenantID, qrPayment.PaymentID, fromAccountID)
	assert.NoError(t, err)

	// Step 4: Inquire about the QR Payment again to verify status change
	inquiredPaymentAfter, err := InquireQRPayment(ctx, _dbSvc, tenantID, qrPayment.PaymentID)
	assert.NoError(t, err)
	assert.NotNil(t, inquiredPaymentAfter)
	assert.Equal(t, "COMPLETED", inquiredPaymentAfter.Status)
}

func TestInquireQRPayment(t *testing.T) {
	type args struct {
		ctx       context.Context
		dbSvc     *dynamodb.Client
		tenantID  string
		paymentID string
	}
	tests := []struct {
		name    string
		args    args
		want    *QRPaymentRequest
		wantErr bool
	}{
		{"inquire about a token", args{context.Background(), _dbSvc, "nil", "2hdIAOPyHPIwb9dUAxphRBx1WaJ"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InquireQRPayment(tt.args.ctx, tt.args.dbSvc, tt.args.tenantID, tt.args.paymentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("InquireQRPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InquireQRPayment() = %v, want %v", got, tt.want)
			}
		})
	}
}
