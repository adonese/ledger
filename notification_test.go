package ledger

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func TestSendEmail(t *testing.T) {
	type args struct {
		sesSvc *ses.Client
		msg    Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test send email", args{sesSvc: _sesSvc, msg: Message{Body: "thiss h a t1ssage", Subject: "teste email", To: "mmbusif@gmail.com"}}, false},
		{"test send email", args{sesSvc: _sesSvc, msg: Message{Body: "this is a test m2essage", Subject: "tesat email", To: "mmbusif@gmail.com"}}, false},
		{"test send email", args{sesSvc: _sesSvc, msg: Message{Body: "this is ae test message", Subject: "test demail", To: "mmbusif@gmail.com"}}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendEmail(tt.args.sesSvc, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
