**Basic Lambda Layer Version**

This example demonstrates how to create a basic Lambda layer version from S3.

```yaml
resources:
  myLayerVersion:
    type: aws/lambda/layerVersion
    spec:
      layerName: my-python-layer
      description: "Basic Python utilities layer"
      content:
        s3Bucket: my-lambda-layers-bucket
        s3Key: python-utils-layer.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
      compatibleArchitectures:
        - x86_64
```