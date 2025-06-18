**Basic Lambda Event Source Mapping**

This example demonstrates how to create a basic event source mapping for an SQS queue.

```yaml
resources:
  orderProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: order-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            console.log('Processing orders:', JSON.stringify(event, null, 2));
            return { statusCode: 200 };
          };

  orderQueueMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.orderProcessorFunction.functionName}
      eventSourceArn: arn:aws:sqs:us-east-1:123456789012:order-queue
      batchSize: 10
      enabled: true
``` 