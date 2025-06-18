**MSK Event Source Mapping**

This example demonstrates how to create an event source mapping for Amazon MSK (Managed Streaming for Kafka).

```yaml
resources:
  kafkaProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: kafka-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.records) {
              console.log('Processing Kafka record:', record);
            }
            return { statusCode: 200 };
          };

  kafkaMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.kafkaProcessorFunction.functionName}
      eventSourceArn: arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster
      topics:
        - user-events
        - order-events
      batchSize: 100
      startingPosition: TRIM_HORIZON
      maximumBatchingWindowInSeconds: 5
      enabled: true
      amazonManagedKafkaEventSourceConfig:
        consumerGroupId: my-consumer-group
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
``` 