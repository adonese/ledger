// this is for creating a lambda function that listens to ddb streams of deleted data

const AWS = require('aws-sdk');
const dynamodb = new AWS.DynamoDB.DocumentClient();

exports.handler = async (event) => {
  const destinationTable = process.env.DESTINATION_TABLE;

  for (const record of event.Records) {
    if (record.eventName === 'REMOVE') {
      const deletedItem = AWS.DynamoDB.Converter.unmarshall(record.dynamodb.OldImage);
      try {
        await dynamodb.put({
          TableName: destinationTable,
          Item: deletedItem
        }).promise();
        console.log(`Moved item with TenantID: ${deletedItem.TenantID} and AccountID: ${deletedItem.AccountID} to archive table.`);
      } catch (err) {
        console.error(`Failed to move item to archive table: ${err}`);
      }
    }
  }
};
