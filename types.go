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
