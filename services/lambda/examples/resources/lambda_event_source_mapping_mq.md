**Amazon MQ Event Source Mapping**

This example demonstrates how to create an event source mapping for Amazon MQ (ActiveMQ/RabbitMQ).

```yaml
resources:
  mqProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: mq-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing MQ message:', record);
            }
            return { statusCode: 200 };
          };

  mqMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.mqProcessorFunction.functionName}
      eventSourceArn: arn:aws:mq:us-east-1:123456789012:broker/my-broker
      queues:
        - order-queue
        - notification-queue
      batchSize: 10
      enabled: true
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
        - type: BASIC_AUTH
          uri: arn:aws:secretsmanager:us-east-1:123456789012:secret:mq-credentials
``` 