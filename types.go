package ledger

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
}
