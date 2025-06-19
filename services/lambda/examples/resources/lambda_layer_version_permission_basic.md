**Basic Lambda Layer Version Permission**

This example demonstrates how to grant layer access permission to a specific AWS account.

```yaml
resources:
  layerPermission:
    type: aws/lambda/layerVersionPermission
    spec:
      layerVersionArn: arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1
      statementId: my-permission
      action: lambda:GetLayerVersion
      principal: "987654321098"
```
