# Complete Lambda Function URL

This example creates a Lambda function URL with all available configuration options.

```yaml
type: aws/lambda/functionUrl
spec:
  targetFunctionArn: my-function
  authType: AWS_IAM
  qualifier: PROD
  invokeMode: RESPONSE_STREAM
  cors:
    allowCredentials: true
    allowHeaders:
      - Content-Type
      - Authorization
    allowMethods:
      - GET
      - POST
      - PUT
      - DELETE
    allowOrigins:
      - https://example.com
      - https://app.example.com
    exposeHeaders:
      - X-Custom-Header
    maxAge: 3600
```
