package ledger

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func ConvertToSDKAttributeValue(av events.DynamoDBAttributeValue) types.AttributeValue {
	switch av.DataType() {
	case events.DataTypeString:
		return &types.AttributeValueMemberS{Value: av.String()}
	case events.DataTypeNumber:
		return &types.AttributeValueMemberN{Value: av.Number()}
	case events.DataTypeBinary:
		return &types.AttributeValueMemberB{Value: av.Binary()}
	case events.DataTypeBoolean:
		return &types.AttributeValueMemberBOOL{Value: av.Boolean()}
	case events.DataTypeNull:
		return &types.AttributeValueMemberNULL{Value: true}
	case events.DataTypeMap:
		m := make(map[string]types.AttributeValue)
		for k, v := range av.Map() {
			m[k] = ConvertToSDKAttributeValue(v)
		}
		return &types.AttributeValueMemberM{Value: m}
	case events.DataTypeList:
		l := make([]types.AttributeValue, len(av.List()))
		for i, v := range av.List() {
			l[i] = ConvertToSDKAttributeValue(v)
		}
		return &types.AttributeValueMemberL{Value: l}
	default:
		return nil
	}
}

func VerifySignature(publicKeyStr, message, signatureStr string) bool {
	// Decode the public key from Base64

	pubKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		log.Fatalf("Failed to decode public key: %v", err)
	}

	// Parse the public key
	pubKey, err := x509.ParsePKIXPublicKey(pubKeyBytes)
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("Public key is not of type *rsa.PublicKey")
	}

	// Decode the signature from Base64
	sigBytes, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		log.Fatalf("Failed to decode signature: %v", err)
	}

	// Create a hash of the message
	hash := sha256.New()
	hash.Write([]byte(message))
	hashed := hash.Sum(nil)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed, sigBytes)
	return err == nil
}
