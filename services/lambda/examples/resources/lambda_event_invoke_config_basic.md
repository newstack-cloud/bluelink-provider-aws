**Basic Lambda Event Invoke Config**

This example demonstrates how to create a basic Event Invoke Config that configures retry settings for a Lambda function.

```yaml
resources:
  myEventInvokeConfig:
    type: aws/lambda/eventInvokeConfig
    spec:
      functionName: my-lambda-function
      qualifier: $LATEST
      maximumRetryAttempts: 1
      maximumEventAgeInSeconds: 300
```