# Basic Lambda Function URL

This example creates a basic Lambda function URL with no authentication required.

```yaml
type: aws/lambda/functionUrl
spec:
  targetFunctionArn: my-function
  authType: NONE
```
