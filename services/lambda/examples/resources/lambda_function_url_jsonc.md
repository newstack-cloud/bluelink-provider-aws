# Lambda Function URL (JSONC)

This example creates a Lambda function URL using JSONC format.

```javascript
{
  "type": "aws/lambda/functionUrl",
  "spec": {
    "targetFunctionArn": "my-function",
    "authType": "NONE",
    "invokeMode": "BUFFERED"
  }
}
```
