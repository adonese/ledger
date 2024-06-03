package ledger

import (
	"testing"
)

var _AWS_REGION = "us-east-1"

func TestInitializeLedger(t *testing.T) {
	type args struct {
		accessKey string
		secretKey string
		region    string
	}
	tests := []struct {
		name string
		args args
	}{
		{"test_initializing aws", args{accessKey: "", secretKey: "", region: _AWS_REGION}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitializeLedger(tt.args.accessKey, tt.args.secretKey, tt.args.region)
			if err != nil {
				t.Errorf("InitializeLedger() error = %v", err)
				return
			}
			if got == nil {
				t.Errorf("InitializeLedger() error = ledger is nil")
				return
			}
		})
	}
}
