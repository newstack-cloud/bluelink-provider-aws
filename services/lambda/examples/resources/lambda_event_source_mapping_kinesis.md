**Kinesis Event Source Mapping**

This example demonstrates how to create an event source mapping for Amazon Kinesis with advanced configurations.

```yaml
resources:
  kinesisProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: kinesis-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing Kinesis record:', record.kinesis.data);
            }
            return { statusCode: 200 };
          };

  kinesisMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.kinesisProcessorFunction.functionName}
      eventSourceArn: arn:aws:kinesis:us-east-1:123456789012:stream/data-stream
      batchSize: 100
      startingPosition: TRIM_HORIZON
      maximumBatchingWindowInSeconds: 5
      maximumRecordAgeInSeconds: 3600
      maximumRetryAttempts: 3
      bisectBatchOnFunctionError: true
      parallelizationFactor: 2
      tumblingWindowInSeconds: 300
      enabled: true
      functionResponseTypes:
        - ReportBatchItemFailures
``` 