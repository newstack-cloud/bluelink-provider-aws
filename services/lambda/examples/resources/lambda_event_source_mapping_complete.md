**Complete Event Source Mapping Example**

This example demonstrates a comprehensive event source mapping with all available configurations.

```yaml
resources:
  comprehensiveProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: comprehensive-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            console.log('Processing event:', JSON.stringify(event, null, 2));
            return { statusCode: 200 };
          };

  comprehensiveMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.comprehensiveProcessorFunction.functionName}
      eventSourceArn: arn:aws:kinesis:us-east-1:123456789012:stream/comprehensive-stream
      batchSize: 200
      startingPosition: AT_TIMESTAMP
      startingPositionTimestamp: 1640995200  # Unix timestamp for 2022-01-01 00:00:00 UTC
      maximumBatchingWindowInSeconds: 15
      maximumRecordAgeInSeconds: 604800  # 7 days
      maximumRetryAttempts: 10
      bisectBatchOnFunctionError: true
      parallelizationFactor: 3
      tumblingWindowInSeconds: 600  # 10 minutes
      enabled: true
      functionResponseTypes:
        - ReportBatchItemFailures
      kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/comprehensive-key
      filterCriteria:
        filters:
          - pattern: '{"source":["aws.kinesis"]}'
          - pattern: '{"detail-type":["DataRecord"]}'
      destinationConfig:
        onSuccess:
          destination: arn:aws:sqs:us-east-1:123456789012:success-queue
        onFailure:
          destination: arn:aws:sqs:us-east-1:123456789012:dlq-queue
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
        - type: VPC_SUBNET
          uri: subnet-87654321
      scalingConfig:
        maximumConcurrency: 100
      provisionedPollerConfig:
        minimumPollers: 5
        maximumPollers: 20
      tags:
        Environment: Production
        Application: DataProcessing
        Owner: DataTeam
``` 