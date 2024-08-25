package ledger

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const SNS_TOPIC = "arn:aws:sns:us-east-1:767397764981:TransactionNotifications"

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
	TenantID          string  `dynamodbav:"TenantID" json:"tenant_id,omitempty"`
	Email             string  `dynamodbav:"Email" json:"email,omitempty"`
}

func NewDefaultAccount(accountId, mobileNumber, name, pubkey, tenantId string) User {
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
		TenantID:          tenantId,
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
	AccountID           string  `dynamodbav:"AccountID" json:"account_id,omitempty"`
	SystemTransactionID string  `dynamodbav:"TransactionID" json:"transaction_id,omitempty"`
	FromAccount         string  `dynamodbav:"FromAccount" json:"from_account,omitempty"`
	ToAccount           string  `dynamodbav:"ToAccount" json:"to_account,omitempty"`
	Amount              float64 `dynamodbav:"Amount" json:"amount"`
	Comment             string  `dynamodbav:"Comment" json:"comment,omitempty"`
	TransactionDate     int64   `dynamodbav:"TransactionDate" json:"time,omitempty"`
	Status              *int    `dynamodbav:"TransactionStatus" json:"status,omitempty"`
	TenantID            string  `dynamodbav:"TenantID" json:"tenant_id,omitempty"`
	InitiatorUUID       string  `dynamodbav:"UUID" json:"uuid,omitempty"`
	Timestamp           string  `dynamodbav:"timestamp" json:"timestamp,omitempty"`
	SignedUUID          string  `dynamodbav:"signed_uuid" json:"signed_uuid,omitempty"`
}

// Create a new transacton entry and populate it with default time and status of 1, using the current time.
// Should we use pointer? or use func (n *TransactionEntry) New() which us better
func NewTransactionEntry(fromAccount, toAccount string, amount float64) TransactionEntry {
	uid := uuid.New().String()
	failedTransaction := 1
	return TransactionEntry{
		SystemTransactionID: uid,
		FromAccount:         fromAccount,
		ToAccount:           toAccount,
		Amount:              amount,
		Comment:             "failed",
		TransactionDate:     getCurrentTimestamp(),
		Status:              &failedTransaction,
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

// NilRresponse
// Status should be: error, success, pending
// Code: a generic nil code message
/*
    "status": "error",
    "code": "insufficient_balance",
    "message": "Insufficient balance to complete the transaction.",
    "details": "The user does not have enough balance in their account.",
    "timestamp": "2024-05-24T12:05:00Z",
    "uuid": "uuid_001",
    "signed_uuid": "signed_uuid_001"
} */

type NilResponse struct {
	Status    string `json:"status,omitempty"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
	Data      data   `json:"data"`
	Details   string `json:"details,omitempty"`
}

type data struct {
	FromAccount   string  `json:"from_account,omitempty"`
	UUID          string  `json:"uuid,omitempty"`
	TransactionID string  `json:"transaction_id,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	SignedUUID    string  `json:"signed_uuid,omitempty"`
	Currency      string  `json:"currency,omitempty"`
}

type Beneficiary struct {
	AccountID  string `dynamodbav:"AccountID" json:"account_id,omitempty"`
	FullName   string `dynamodbav:"FullName" json:"full_name,omitempty"`
	Mobile     string `dynamodbav:"Mobile" json:"mobile,omitempty"`
	Provider   string `dynamodbav:"Provider" json:"provider,omitempty"`
	Address    string `dynamodbav:"Address" json:"address,omitempty"`
	BranchName string `dynamodbav:"BranchName" json:"branch_name,omitempty"`
}
type EscrowTransaction struct {
	SystemTransactionID string      `dynamodbav:"TransactionID" json:"transaction_id,omitempty"`
	FromAccount         string      `dynamodbav:"FromAccount" json:"from_account,omitempty"`
	ToAccount           string      `dynamodbav:"ToAccount" json:"to_account,omitempty"`
	Amount              float64     `dynamodbav:"Amount" json:"amount"`
	Comment             string      `dynamodbav:"Comment" json:"comment,omitempty"`
	TransactionDate     int64       `dynamodbav:"TransactionDate" json:"time,omitempty"`
	Status              Status      `dynamodbav:"TransactionStatus" json:"status,omitempty"`
	FromTenantID        string      `dynamodbav:"FromTenantID" json:"from_tenant_id,omitempty"`
	ToTenantID          string      `dynamodbav:"ToTenantID" json:"to_tenant_id,omitempty"`
	InitiatorUUID       string      `dynamodbav:"UUID" json:"uuid,omitempty"`
	Timestamp           string      `dynamodbav:"timestamp" json:"timestamp,omitempty"`
	SignedUUID          string      `dynamodbav:"signed_uuid" json:"signed_uuid,omitempty"`
	CashoutProvider     string      `dynamodbav:"CashoutProvider" json:"cashout_provider,omitempty"`
	Beneficiary         Beneficiary `dynamodbav:"Beneficiary" json:"beneficiary,omitempty"`
	TransientAccount    string      `dynamodbav:"TransientAccount" json:"transient_account,omitempty"`
	TransientTenant     string      `dynamodbav:"TransientTenant" json:"transient_tenant,omitempty"`
	ServiceProvider     string      `dynamodbav:"ServiceProvider" json:"service_provider,omitempty"`
}

type EscrowMeta struct {
	TenantID           string   `dynamodbav:"TenantID" json:"from_tenant_id,omitempty"`
	Webhook            string   `dynamodbav:"Webhook" json:"webhook,omitempty"`
	AllowedTenants     []string `dynamodbav:"AllowedTenants" json:"allowed_tenants,omitempty"`
	Currency           string   `dynamodbav:"Currency" json:"currency,omitempty"`
	SupportsConversion bool     `dynamodbav:"SupportsConversion" json:"supports_conversion,omitempty"`
}

type EscrowEntry struct {
	FromAccount       string      `dynamodbav:"FromAccount" json:"from_account,omitempty"`
	ToAccount         string      `dynamodbav:"ToAccount" json:"to_account,omitempty"`
	Amount            float64     `dynamodbav:"Amount" json:"amount"`
	Comment           string      `dynamodbav:"Comment" json:"comment"`
	NotEscrowTenantID string      `dynamodbav:"TenantID" json:"tenant_id,omitempty"`
	InitiatorUUID     string      `dynamodbav:"UUID" json:"uuid,omitempty"`
	Timestamp         string      `dynamodbav:"timestamp" json:"timestamp"`
	SignedUUID        string      `dynamodbav:"signed_uuid" json:"signed_uuid,omitempty"`
	ToTenantID        string      `dynamodbav:"ToTenantID" json:"to_tenant_id,omitempty"`
	FromTenantID      string      `dynamodbav:"FromTenantID" json:"from_tenant_id,omitempty"`
	CashoutProvider   string      `dynamodbav:"CashoutProvider" json:"cashout_provider"`
	Beneficiary       Beneficiary `dynamodbav:"Beneficiary" json:"beneficiary"`
	ServiceProvider   string      `dynamodbav:"ServiceProvider" json:"service_provider"`
}

type ServiceProvider struct {
	TenantID      string `dynamodbav:"TenantID" json:"tenant_id"`
	WebhookURL    string `dynamodbav:"WebhookURL" json:"webhook_url"`
	TailscaleURL  string `dynamodbav:"TailscaleURL" json:"tailscale_url"`
	LastAccessed  string `dynamodbav:"LastAccessed" json:"last_accessed"`
	Currency      string `dynamodbav:"Currency" json:"currency"`
	PublicKey     string `dynamodbav:"PublicKey" json:"public_key"`
	Email         string `dynamodbav:"Email" json:"email"`
	EscrowAccount string `dynamodbav:"EscrowAccount" json:"escrow_account"`
}

// Status represents the status of a transaction
type Status int

// Define possible statuses
const (
	StatusPending Status = iota
	StatusCompleted
	StatusFailed
	StatusInProgress
)

// Map from string to Status
var statusStringToEnum = map[string]Status{
	"Pending":    StatusPending,
	"Completed":  StatusCompleted,
	"Failed":     StatusFailed,
	"InProgress": StatusInProgress,
}

// Map from Status to string (optional, for marshalling)
var statusEnumToString = map[Status]string{
	StatusPending:    "Pending",
	StatusCompleted:  "Completed",
	StatusFailed:     "Failed",
	StatusInProgress: "InProgress",
}

// UnmarshalDynamoDBAttributeValue implements custom unmarshalling for Status
func (s *Status) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		if status, ok := statusStringToEnum[v.Value]; ok {
			*s = status
			return nil
		}
		return fmt.Errorf("unknown status string: %s", v.Value)
	case *types.AttributeValueMemberN:
		i, err := strconv.Atoi(v.Value)
		if err != nil {
			return fmt.Errorf("failed to parse status number: %v", err)
		}
		*s = Status(i)
		return nil
	default:
		return fmt.Errorf("attribute value is not a string or number")
	}
}

// String returns the string representation of the Status
func (s Status) String() string {
	if str, ok := statusEnumToString[s]; ok {
		return str
	}
	return fmt.Sprintf("UnknownStatus(%d)", s)
}

type QueryResultEscrowWebhookTable struct {
	Transactions     []EscrowTransaction
	LastEvaluatedKey map[string]types.AttributeValue
	HasMorePages     bool
}
