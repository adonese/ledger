package main

import (
	"log"
	"os"
	"testing"

	"github.com/adonese/ledger"
)

func Test_sign(t *testing.T) {
	type args struct {
		data       string
		privateKey []byte
	}

	// read privatekey from file priv.pem
	privKey, err := os.ReadFile("priv.pem")
	if err != nil {
		t.Fatalf("Failed to read private key: %v", err)
	}

	log.Printf("the private key is: %s", privKey)
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"test1", args{"test", privKey}, "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sign(tt.args.data, tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !ledger.VerifySignature("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA6n9XrRSSZZM46mmsE3F0qVjnFgcGKySy+jaTuOX2QjNI8qysbyL/hoDqhYhmOoPPbwn18JO2Ochw+EXcbKnR9qAPIu8CEeUweo0LG+Cv5SL/WBI2kaWpDz3fMSzw+Hanf6hRqm7jsWR/RV5qPI73IdBJ3gfdUpv9Ta8uzk7HOwIuR30Ja7pLKleIf5HFt56uFx8dxAofv7I8cc0NFbhKa7A937/DyqQG7vE+CGlF2MZPdMw0HMfOCxFWGekVwlrwkmdxjgtaNYJrtxHmzHOwVcnT7/7kGZrZ5GxefuV6eMo2ed4y0/QF/wzyZuBCQATkL962xiELcGkjzIIbcb1YlQIDAQAB", tt.args.data, got) {
				t.Errorf("VerifySignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
