package ledger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"testing"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type _ledgerSecrets struct {
	AccessKey string `json:"AWS_ACCESS_KEY_ID"`
	SecretKey string `json:"AWS_SECRET_ACCESS_KEY"`
}

//go:embed .secrets.json
var secrets []byte

var ledgerSecret _ledgerSecrets

func init() {
	if err := json.Unmarshal(secrets, &ledgerSecret); err != nil {
		log.Printf("the error is: %v", err)
		os.Exit(1)
	}
}

var _dbSvc *dynamodb.Client
var _sesSvc *ses.Client

func init() {
	var err error

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(_AWS_REGION),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			ledgerSecret.AccessKey,
			ledgerSecret.SecretKey,
			"",
		)),
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
		tenantId      string
		context       context.Context
	}
	ctx := context.TODO()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		// 121336038
		// 6224
		// 6224
		// there's a severe bug in transfering credits
		{"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 67, tenantId: "zero", context: ctx}, false},
		{"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", dbSvc: _dbSvc, amount: 10000, tenantId: "nil", context: ctx}, false},
		// {"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", dbSvc: _dbSvc, amount: 10000, tenantId: "noooon"}, true},
		// {"testing transfer", args{fromAccountID: "0111493885", toAccountID: "151515", dbSvc: _dbSvc, amount: 323222121}, false},
		// {"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "12", dbSvc: _dbSvc, amount: 151}, false},
		// {"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "12", dbSvc: _dbSvc, amount: 120}, false},
		// {"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "12", dbSvc: _dbSvc, amount: 32}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "249_ACCT_1", dbSvc: _dbSvc, amount: 43}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "249_ACCT_1", dbSvc: _dbSvc, amount: 324}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "249_ACCT_1", dbSvc: _dbSvc, amount: 1210}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "249_ACCT_1", dbSvc: _dbSvc, amount: 322}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 43}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 324}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 1210}, false},
		// {"testing transfer", args{fromAccountID: "12", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 322}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trEntry := TransactionEntry{TenantID: tt.args.tenantId, FromAccount: tt.args.fromAccountID, ToAccount: tt.args.toAccountID,
				Amount: tt.args.amount, AccountID: tt.args.fromAccountID, InitiatorUUID: uuid.NewString()}
			res, err := TransferCredits(tt.args.context, tt.args.dbSvc, trEntry)
			if err != nil {
				t.Errorf("transferCredits() error = %v, wantErr %v", err, tt.wantErr)
			}
			if res.Code != "failed" {
				t.Errorf("transferCredits() error = %+v", res)
			}

		})
	}
}

func Test_inquireBalance(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		AccountID string
		tenantId  string
		context   context.Context
	}
	ctx := context.TODO()
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "0111498888", tenantId: "", context: ctx}, 30, false},
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "249_ACCT_1", tenantId: "", context: ctx}, 2636, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InquireBalance(tt.args.context, tt.args.dbSvc, "", tt.args.AccountID)
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
		tenantId  string
		context   context.Context
	}
	ctx := context.TODO()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "249_ACCT_1", amount: 121342212, tenantId: "zero", context: ctx}, true},
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "0111493885", amount: 500, tenantId: "othernil", context: ctx}, true},
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "0111493885", amount: 500, tenantId: "nonil", context: ctx}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccountWithBalance(tt.args.context, tt.args.dbSvc, tt.args.tenantId, tt.args.accountId, (tt.args.amount)); err != nil {
				t.Errorf("createAccountWithBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckUser(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		accountId string
		tenantId  string
		context   context.Context
	}

	ctx := context.TODO()

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "0111493885", tenantId: "nil", context: ctx}, true, true},
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "0111493885", tenantId: "unil", context: ctx}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notF, err := CheckUsersExist(tt.args.context, tt.args.dbSvc, tt.args.tenantId, []string{tt.args.accountId})
			if err != nil {
				t.Errorf("there's an error: %v - notfound: %v", err, notF)
				return
			}
		})
	}
}

type MockDynamoDBClient struct {
	mock.Mock
	dynamodb.Client
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestGetTransactions(t *testing.T) {
	type args struct {
		dbSvc            *dynamodb.Client
		accountID        string
		limit            int32
		lastEvaluatedKey string
		context          context.Context
	}
	ctx := context.TODO()
	tests := []struct {
		name    string
		args    args
		want    []LedgerEntry
		want1   string
		wantErr bool
	}{
		{"test-retrieving results", args{context: ctx, dbSvc: _dbSvc, accountID: "249_ACCT_1", limit: 2, lastEvaluatedKey: ""}, []LedgerEntry{{}}, "12345", false},
		{"test-retrieving results", args{context: ctx, dbSvc: _dbSvc, accountID: "249_ACCT_1", limit: 2, lastEvaluatedKey: "62fadf6c-5f4a-441a-865a-34b84a49040f"}, []LedgerEntry{{}}, "12345", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetTransactions(tt.args.context, tt.args.dbSvc, "nil", tt.args.accountID, tt.args.limit, tt.args.lastEvaluatedKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTransactions() error = %v, wantErr %v - and key is: %s", err, tt.wantErr, got1)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactions() got = %v, want %v - and key is: %v", got, tt.want, got1)
			}
			if got1 != tt.want1 {
				t.Errorf("GetTransactions() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetDetailedTransactions(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		accountID string
		limit     int32
		context   context.Context
	}

	ctx := context.TODO()

	tests := []struct {
		name    string
		args    args
		want    []TransactionEntry
		wantErr bool
	}{
		{
			name: "Fetch transactions for 249_ACCT_1",
			args: args{
				dbSvc:     _dbSvc, // Assuming _dbSvc is your DynamoDB client
				accountID: "0111493885",
				limit:     4,
				context:   ctx,
			},
			want: []TransactionEntry{
				{
					AccountID:           "249_ACCT_1",
					SystemTransactionID: "tx1", // Replace with the actual TransactionID
					FromAccount:         "249_ACCT_1",
					ToAccount:           "12",
					Amount:              10,
					Comment:             "Transfer credits",
					TransactionDate:     1632835600, // Replace with the actual TransactionDate
				},
				{
					AccountID:           "249_ACCT_1",
					SystemTransactionID: "tx2", // Replace with the actual TransactionID
					FromAccount:         "249_ACCT_1",
					ToAccount:           "12",
					Amount:              15,
					Comment:             "Transfer credits",
					TransactionDate:     1632835600, // Replace with the actual TransactionDate
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDetailedTransactions(tt.args.context, tt.args.dbSvc, "nil", tt.args.accountID, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDetailedTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDetailedTransactions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGetAllNilTransactions(t *testing.T) {
	ctx := context.TODO()
	tests := []struct {
		name     string
		filter   TransactionFilter
		wantMin  int
		tenantId string
		context  context.Context
	}{
		{
			name:     "Fetch all transactions",
			filter:   TransactionFilter{},
			wantMin:  28, // Adjust based on expected data in your test table
			tenantId: "nil",
			context:  ctx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := GetAllNilTransactions(ctx, _dbSvc, tt.tenantId, tt.filter)
			if err != nil {
				t.Errorf("GetAllNilTransactions() error = %v", err)
				return
			}
			if len(got) != tt.wantMin {
				t.Errorf("GetAllNilTransactions() got %v, want at least %v results - the result is: %+v", len(got), tt.wantMin, got)
			}
		})
	}
}

// SetupTestData initializes the DynamoDB table with predefined accounts and balances.
func setupTestData(dbSvc *dynamodb.Client) {
	accounts := []UserBalance{
		{"0111493885", 500},
		{"0111493885", 6224},
		{"0111493888", 0},
		{"0111498888", 0},
		{"249_ACCT_1", 121336038},
		{"249_ACCT_1", 121341183},
		{"0111493885", 500},
	}

	for _, acc := range accounts {
		item, _ := attributevalue.MarshalMap(acc)
		dbSvc.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(NilUsers),
			Item:      item,
		})
	}
}

func TestTransferCredits(t *testing.T) {
	dbSvc := _dbSvc
	setupTestData(dbSvc)

	type args struct {
		dbSvc         *dynamodb.Client
		fromAccountID string
		toAccountID   string
		amount        float64
		tenantId      string
		context       context.Context
	}
	ctx := context.TODO()
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		expectedCode string
		beforeFrom   float64
		beforeTo     float64
		afterFrom    float64
		afterTo      float64
	}{
		{
			name:         "Basic Transfer",
			args:         args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", amount: 10000, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      false,
			expectedCode: "successful_transaction",
			beforeFrom:   121336038,
			beforeTo:     0,
			afterFrom:    121326038,
			afterTo:      10000,
		},
		{
			name:         "Insufficient Funds",
			args:         args{fromAccountID: "0111493888", toAccountID: "0111498888", amount: 1, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      true,
			expectedCode: "insufficient_balance",
			beforeFrom:   0,
			beforeTo:     0,
			afterFrom:    0,
			afterTo:      0,
		},
		{
			name:         "Non-existent Sender",
			args:         args{fromAccountID: "nonexistent", toAccountID: "0111498888", amount: 1, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      true,
			expectedCode: "user_not_found",
			beforeFrom:   0,
			beforeTo:     0,
			afterFrom:    0,
			afterTo:      0,
		},
		{
			name:         "Non-existent Receiver",
			args:         args{fromAccountID: "249_ACCT_1", toAccountID: "nonexistent", amount: 1000, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      true,
			expectedCode: "user_not_found",
			beforeFrom:   121336038,
			beforeTo:     0,
			afterFrom:    121336038,
			afterTo:      0,
		},
		{
			name:         "Zero Transfer",
			args:         args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", amount: 0, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      true,
			expectedCode: "invalid_amount",
			beforeFrom:   121336038,
			beforeTo:     0,
			afterFrom:    121336038,
			afterTo:      0,
		},
		{
			name:         "Negative Transfer",
			args:         args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", amount: -100, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      true,
			expectedCode: "invalid_amount",
			beforeFrom:   121336038,
			beforeTo:     0,
			afterFrom:    121336038,
			afterTo:      0,
		},
		{
			name:         "Floating Point Transfer",
			args:         args{fromAccountID: "249_ACCT_1", toAccountID: "0111493888", amount: 1234.56, tenantId: "nil", dbSvc: dbSvc, context: ctx},
			wantErr:      false,
			expectedCode: "successful_transaction",
			beforeFrom:   121336038,
			beforeTo:     0,
			afterFrom:    121334803.44,
			afterTo:      1234.56,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fromAccountBalance, _ := InquireBalance(tt.args.context, tt.args.dbSvc, tt.args.tenantId, tt.args.fromAccountID)
			toAccountBalance, _ := InquireBalance(tt.args.context, tt.args.dbSvc, tt.args.tenantId, tt.args.toAccountID)

			assert.Equal(t, tt.beforeFrom, fromAccountBalance, "Before fromAccount balance should match")
			assert.Equal(t, tt.beforeTo, toAccountBalance, "Before toAccount balance should match")

			trEntry := TransactionEntry{
				TenantID:      tt.args.tenantId,
				FromAccount:   tt.args.fromAccountID,
				ToAccount:     tt.args.toAccountID,
				Amount:        tt.args.amount,
				AccountID:     tt.args.fromAccountID,
				InitiatorUUID: uuid.NewString(),
			}
			res, err := TransferCredits(tt.args.context, tt.args.dbSvc, trEntry)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferCredits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if res.Code != tt.expectedCode {
				t.Errorf("TransferCredits() = %+v, expectedCode %v", res, tt.expectedCode)
			}

			// Capture balances after transfer
			fromAccountBalanceAfter, _ := InquireBalance(tt.args.context, tt.args.dbSvc, tt.args.tenantId, tt.args.fromAccountID)
			toAccountBalanceAfter, _ := InquireBalance(tt.args.context, tt.args.dbSvc, tt.args.tenantId, tt.args.toAccountID)

			assert.Equal(t, tt.afterFrom, fromAccountBalanceAfter, "After fromAccount balance should match")
			assert.Equal(t, tt.afterTo, toAccountBalanceAfter, "After toAccount balance should match")
		})
	}
}
