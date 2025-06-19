**Lambda Layer Version Permission JSONC Example**

This example demonstrates how to grant layer access permission to a specific AWS account using JSONC format.

```javascript
{
  "resources": {
    "layerPermission": {
      "type": "aws/lambda/layerVersionPermission",
      "spec": {
        "layerVersionArn": "arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
        "statementId": "my-permission",
        "action": "lambda:GetLayerVersion",
        "principal": "987654321098"
      }
    }
  }
}
```
