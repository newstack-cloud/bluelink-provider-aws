**Complete Lambda Layer Version Permission Example**

This example demonstrates how to create a complete Lambda setup with a layer, layer version, and organization-wide permissions with all available options.

```yaml
resources:
  myLayer:
    type: aws/lambda/layerVersion
    spec:
      layerName: my-shared-utilities
      description: "Shared utilities layer for organization"
      content:
        s3Bucket: my-lambda-layers-bucket
        s3Key: utilities-layer.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - nodejs18.x
      compatibleArchitectures:
        - x86_64
        - arm64

  organizationPermission:
    type: aws/lambda/layerVersionPermission
    spec:
      layerVersionArn: ${resources.myLayer.layerVersionArn}
      statementId: organization-access
      action: lambda:GetLayerVersion
      principal: "*"
      organizationId: o-abc123defg
```