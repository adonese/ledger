package ledger

import (
	"encoding/json"
	"testing"
)

func TestNewDefaultAccount(t *testing.T) {

	tests := []struct {
		name string
		args string
		want User
	}{
		{"test-new-account", `{"fullname": "test", "mobile": "123456789", "account_id": "249_ACCT_1", "user_pubkey": "123456789"}`, User{MobileNumber: "123456789", FullName: "test", AccountID: "249_ACCT_1", PublicKey: "123456789"}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var user User
			json.Unmarshal([]byte(tt.args), &user)
			if user.FullName != tt.want.FullName {
				t.Errorf("NewDefaultAccount() = %v, want %v", user.FullName, tt.want.FullName)
			}
			if user.MobileNumber != tt.want.MobileNumber {
				t.Errorf("NewDefaultAccount() = %v, want %v", user.MobileNumber, tt.want.MobileNumber)
			}
			if user.AccountID != tt.want.AccountID {
				t.Errorf("NewDefaultAccount() = %v, want %v", user.AccountID, tt.want.AccountID)
			}
			if user.PublicKey != tt.want.PublicKey {
				t.Errorf("NewDefaultAccount() = %v, want %v", user.PublicKey, tt.want.PublicKey)
			}

		})
	}
}
