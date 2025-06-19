**Complete Lambda Event Invoke Config**

This example demonstrates how to create a complete Event Invoke Config with retry settings and destinations for success and failure.

```yaml
resources:
  myEventInvokeConfig:
    type: aws/lambda/eventInvokeConfig
    spec:
      functionName: my-lambda-function
      qualifier: $LATEST
      maximumRetryAttempts: 2
      maximumEventAgeInSeconds: 1800
      destinationConfig:
        onSuccess:
          destination: arn:aws:sqs:us-east-1:123456789012:success-queue
        onFailure:
          destination: arn:aws:sqs:us-east-1:123456789012:failure-queue
```
