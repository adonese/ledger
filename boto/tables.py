import boto3

# Initialize a DynamoDB client
dynamodb = boto3.resource('dynamodb')

# Specify your table name
table_name = 'TransactionsTable'

# Get the table
table = dynamodb.Table(table_name)

# Scan the table and get all items
response = table.scan()
items = response['Items']

# Iterate over all items and update each one
for item in items:
    table.update_item(
        Key={'TransactionID': item['TransactionID']},
        UpdateExpression='SET TransactionStatus = :val',
        ExpressionAttributeValues={':val': 0}
    )