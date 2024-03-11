package ledger

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// SMS defines the structure for sending SMS notifications.
// It includes the API key, sender information, recipient mobile number, message content,
// and the SMS gateway URL.
type SMS struct {
	APIKey  string `json:"api_key,omitempty"`
	Sender  string `json:"sender,omitempty"`
	Mobile  string `json:"mobile,omitempty"`
	Message string `json:"message,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

// Message defines the structure for sending email messages.
// It includes the message body, subject, and recipient address.
type Message struct {
	Body    string `json:"body,omitempty"`
	Subject string `json:"subject,omitempty"`
	To      string `json:"to,omitempty"`
}

// User represent a user entry in nil users table.
type User struct {
	AccountID         string  `dynamodbav:"AccountID" json:"account_id,omitempty"`
	FullName          string  `dynamodbav:"full_name" json:"full_name,omitempty"`
	Birthday          string  `dynamodbav:"birthday" json:"birthday,omitempty"`
	City              string  `dynamodbav:"city" json:"city,omitempty"`
	Dependants        int     `dynamodbav:"dependants" json:"dependants,omitempty"`
	IncomeLastYear    float64 `dynamodbav:"income_last_year" json:"income_last_year,omitempty"`
	EnrollSMEsProgram bool    `dynamodbav:"enroll_smes_program" json:"enroll_smes_program,omitempty"`
	Confirm           bool    `dynamodbav:"confirm" json:"confirm,omitempty"`
	ExternalAuth      bool    `dynamodbav:"external_auth" json:"external_auth,omitempty"`
	Password          string  `dynamodbav:"password" json:"password,omitempty"`
	CreatedAt         string  `dynamodbav:"created_at" json:"created_at,omitempty"`
	IsVerified        bool    `dynamodbav:"is_verified" json:"is_verified,omitempty"`
	IDType            string  `dynamodbav:"id_type" json:"id_type,omitempty"`
	MobileNumber      string  `dynamodbav:"mobile_number" json:"mobile_number,omitempty"`
	IDNumber          string  `dynamodbav:"id_number" json:"id_number,omitempty"`
	PicIDCard         string  `dynamodbav:"pic_id_card" json:"pic_id_card,omitempty"`
	Amount            float64 `dynamodbav:"amount" json:"amount,omitempty"`
	Currency          string  `dynamodbav:"currency" json:"currency,omitempty"`
	Version           int64   `dynamodbav:"Version" json:"version,omitempty"`
	PublicKey         string  `json:"public_key,omitempty"`
}

func NewDefaultAccount(accountId, mobileNumber, name, pubkey string) User {
	return User{
		AccountID:         accountId,
		FullName:          name,
		Birthday:          "",
		City:              "",
		Dependants:        0,
		IncomeLastYear:    0,
		EnrollSMEsProgram: false,
		Confirm:           false,
		ExternalAuth:      false,
		Password:          "",
		CreatedAt:         time.Now().Local().String(),
		IsVerified:        true,
		IDType:            "",
		MobileNumber:      mobileNumber,
		IDNumber:          "",
		PicIDCard:         "",
		Amount:            0,
		Currency:          "SDG",
	}
}

func (u *User) UnmarshalJSON(b []byte) error {
	type Alias User
	aux := &struct {
		Mobile    string `json:"mobile"`
		FullName  string `json:"fullname"`
		PublicKey string `json:"user_pubkey"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	u.MobileNumber = aux.Mobile // map "mobile" to "MobileNumber"
	u.FullName = aux.FullName   // map "fullname" to "full_name"
	u.PublicKey = aux.PublicKey // map "user_pubkey" to "public_key"
	return nil
}

type TransactionEntry struct {
	AccountID       string  `dynamodbav:"AccountID" json:"account_id,omitempty"`
	TransactionID   string  `dynamodbav:"TransactionID" json:"transaction_id,omitempty"`
	FromAccount     string  `dynamodbav:"FromAccount" json:"from_account,omitempty"`
	ToAccount       string  `dynamodbav:"ToAccount" json:"to_account,omitempty"`
	Amount          float64 `dynamodbav:"Amount" json:"amount,omitempty"`
	Comment         string  `dynamodbav:"Comment" json:"comment,omitempty"`
	TransactionDate int64   `dynamodbav:"TransactionDate" json:"time,omitempty"`
	Status          *int    `dynamodbav:"TransactionStatus" json:"status,omitempty"`
}

// Create a new transacton entry and populate it with default time and status of 1, using the current time. Should we use pointer? or use func (n *TransactionEntry) New() which us better
func NewTransactionEntry(fromAccount, toAccount string, amount float64) TransactionEntry {
	uid := uuid.New().String()
	failedTransaction := 1
	return TransactionEntry{
		TransactionID:   uid,
		FromAccount:     fromAccount,
		ToAccount:       toAccount,
		Amount:          amount,
		Comment:         "failed",
		TransactionDate: getCurrentTimestamp(),
		Status:          &failedTransaction,
	}
}

type TransactionFilter struct {
	AccountID         string
	TransactionStatus *int
	StartTime         int64
	EndTime           int64
	LastEvaluatedKey  map[string]types.AttributeValue
	Limit             int32
}
