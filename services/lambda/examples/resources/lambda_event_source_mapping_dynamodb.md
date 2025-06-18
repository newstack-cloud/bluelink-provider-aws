**DynamoDB Streams Event Source Mapping**

This example demonstrates how to create an event source mapping for DynamoDB Streams.

```yaml
resources:
  dynamodbProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: dynamodb-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing DynamoDB change:', record.dynamodb);
            }
            return { statusCode: 200 };
          };

  dynamodbMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.dynamodbProcessorFunction.functionName}
      eventSourceArn: arn:aws:dynamodb:us-east-1:123456789012:table/users/stream/2024-01-01T00:00:00.000
      batchSize: 50
      startingPosition: LATEST
      maximumBatchingWindowInSeconds: 10
      maximumRecordAgeInSeconds: 7200
      maximumRetryAttempts: 5
      bisectBatchOnFunctionError: true
      parallelizationFactor: 1
      enabled: true
      functionResponseTypes:
        - ReportBatchItemFailures
``` 