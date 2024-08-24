package ledger

import (
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
