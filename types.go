package ledger

type SMS struct {
	APIKey  string `json:"api_key,omitempty"`
	Sender  string `json:"sender,omitempty"`
	Mobile  string `json:"mobile,omitempty"`
	Message string `json:"message,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type Message struct {
	Body    string `json:"body,omitempty"`
	Subject string `json:"subject,omitempty"`
	To      string `json:"to,omitempty"`
}
