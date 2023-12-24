package ledger

import (
	"context"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/stretchr/testify/mock"
)

var _dbSvc *dynamodb.Client
var _sesSvc *ses.Client

func init() {
	var err error

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(_AWS_REGION),
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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		// {"testing transfer", args{fromAccountID: "249_ACCT_1", toAccountID: "0111493885", dbSvc: _dbSvc, amount: 101}, false},
		{"testing transfer", args{fromAccountID: "0111493885", toAccountID: "01114343493885", dbSvc: _dbSvc, amount: 10}, false},
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
			if err := TransferCredits(tt.args.dbSvc, tt.args.fromAccountID, tt.args.toAccountID, tt.args.amount); (err != nil) != tt.wantErr {
				t.Errorf("transferCredits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_inquireBalance(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		AccountID string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "249_ACCT_1"}, 30, false},
		{"test-get-balance", args{dbSvc: _dbSvc, AccountID: "12"}, 2636, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InquireBalance(tt.args.dbSvc, tt.args.AccountID)
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
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"generate account with balance", args{dbSvc: _dbSvc, accountId: "249_ACCT_1", amount: 121342212}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccountWithBalance(tt.args.dbSvc, tt.args.accountId, (tt.args.amount)); err != nil {
				t.Errorf("createAccountWithBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckUser(t *testing.T) {
	type args struct {
		dbSvc     *dynamodb.Client
		accountId string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"testIsUser", args{dbSvc: _dbSvc, accountId: "adonese"}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notF, err := CheckUsersExist(tt.args.dbSvc, []string{tt.args.accountId, "44322"})
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
			got, got1, err := GetTransactions(tt.args.dbSvc, tt.args.accountID, tt.args.limit, tt.args.lastEvaluatedKey)
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
				accountID: "249_ACCT_1",
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
			got, err := GetDetailedTransactions(tt.args.dbSvc, tt.args.accountID, tt.args.limit)
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
