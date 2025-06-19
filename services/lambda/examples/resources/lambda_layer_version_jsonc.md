**Lambda Layer Version (JSONC)**

This example demonstrates how to create a Lambda layer version using JSONC format.

```javascript
{
  "resources": {
    "myLayerVersion": {
      "type": "aws/lambda/layerVersion",
      "spec": {
        "layerName": "my-nodejs-layer",
        "description": "Node.js utilities and dependencies",
        "content": {
          "s3Bucket": "my-lambda-layers-bucket",
          "s3Key": "nodejs-utils-layer.zip",
          "s3ObjectVersion": "abc123def456"
        },
        "compatibleRuntimes": [
          "nodejs18.x",
          "nodejs20.x",
          "nodejs22.x"
        ],
        "compatibleArchitectures": [
          "x86_64",
          "arm64"
        ],
        "licenseInfo": "MIT"
      }
    }
  }
}