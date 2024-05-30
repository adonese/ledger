package ledger

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"testing"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 1029, tenantId: "zero"}, false},
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
			trEntry := TransactionEntry{TenantID: tt.args.tenantId, FromAccount: tt.args.fromAccountID, ToAccount: tt.args.fromAccountID, Amount: tt.args.amount, AccountID: tt.args.fromAccountID}
			if err := TransferCredits(tt.args.dbSvc, trEntry); (err != nil) != tt.wantErr {
				t.Errorf("transferCredits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_inquireBalance(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		AccountID string
		tenantId  string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "0111493885", tenantId: ""}, 30, false},
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "249_ACCT_1", tenantId: ""}, 2636, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InquireBalance(tt.args.dbSvc, "", tt.args.AccountID)
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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "249_ACCT_1", amount: 121342212, tenantId: "zero"}, true},
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "0111493885", amount: 500, tenantId: "othernil"}, true},
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "0111493885", amount: 500, tenantId: "nonil"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccountWithBalance(tt.args.dbSvc, tt.args.tenantId, tt.args.accountId, (tt.args.amount)); err != nil {
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
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "0111493885", tenantId: "nil"}, true, true},
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "0111493885", tenantId: "unil"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notF, err := CheckUsersExist(tt.args.dbSvc, tt.args.tenantId, []string{tt.args.accountId})
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
	}
	tests := []struct {
		name    string
		args    args
		want    []LedgerEntry
		want1   string
		wantErr bool
	}{
		{"test-retrieving results", args{dbSvc: _dbSvc, accountID: "249_ACCT_1", limit: 2, lastEvaluatedKey: ""}, []LedgerEntry{{}}, "12345", false},
		{"test-retrieving results", args{dbSvc: _dbSvc, accountID: "249_ACCT_1", limit: 2, lastEvaluatedKey: "62fadf6c-5f4a-441a-865a-34b84a49040f"}, []LedgerEntry{{}}, "12345", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetTransactions(tt.args.dbSvc, "nil", tt.args.accountID, tt.args.limit, tt.args.lastEvaluatedKey)
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
	}
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
			},
			want: []TransactionEntry{
				{
					AccountID:       "249_ACCT_1",
					TransactionID:   "tx1", // Replace with the actual TransactionID
					FromAccount:     "249_ACCT_1",
					ToAccount:       "12",
					Amount:          10,
					Comment:         "Transfer credits",
					TransactionDate: 1632835600, // Replace with the actual TransactionDate
				},
				{
					AccountID:       "249_ACCT_1",
					TransactionID:   "tx2", // Replace with the actual TransactionID
					FromAccount:     "249_ACCT_1",
					ToAccount:       "12",
					Amount:          15,
					Comment:         "Transfer credits",
					TransactionDate: 1632835600, // Replace with the actual TransactionDate
				},
				// Add the rest of the transactions here...
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDetailedTransactions(tt.args.dbSvc, "nil", tt.args.accountID, tt.args.limit)
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

	// transactionStatus := 1
	// Define your tests
	tests := []struct {
		name     string
		filter   TransactionFilter
		wantMin  int // Use wantMin to specify the minimum number of results expected
		tenantId string
	}{
		{
			name:     "Fetch all transactions",
			filter:   TransactionFilter{},
			wantMin:  28, // Adjust based on expected data in your test table
			tenantId: "nil",
		},
		// {
		// 	name: "Fetch transactions for specific account",
		// 	filter: TransactionFilter{
		// 		AccountID: "exampleAccountID", // Adjust to an existing account ID in your test data
		// 		Limit:     50,
		// 	},
		// 	wantMin: 1, // Ensure this account has at least one transaction
		// },
		// {
		// 	name: "Fetch transactions with status",
		// 	filter: TransactionFilter{
		// 		TransactionStatus: &transactionStatus, // Assuming 0 represents a specific status in your data model
		// 		Limit:             50,
		// 	},
		// 	wantMin: 1,
		// },
	}

	ctx := context.TODO()

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
