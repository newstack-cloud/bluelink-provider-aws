**Complete Lambda Layer Version**

This example demonstrates how to create Lambda layer versions with different deployment options.

```yaml
resources:
  # Layer version from S3 with all optional fields
  fullS3LayerVersion:
    type: aws/lambda/layerVersion
    spec:
      layerName: comprehensive-python-layer
      description: "Comprehensive Python layer with data science libraries"
      content:
        s3Bucket: my-company-lambda-layers
        s3Key: data-science-layer-v2.3.1.zip
        s3ObjectVersion: "version-12345abcdef"
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
        - python3.12
      compatibleArchitectures:
        - x86_64
        - arm64
      licenseInfo: "Apache-2.0"

  # Multi-runtime layer for cross-platform utilities
  universalLayerVersion:
    type: aws/lambda/layerVersion
    spec:
      layerName: universal-utilities
      description: "Universal utilities compatible with multiple runtimes"
      content:
        s3Bucket: shared-lambda-layers
        s3Key: universal-utils/latest.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
        - nodejs18.x
        - nodejs20.x
        - java11
        - java17
        - java21
        - dotnet6
        - dotnet8
        - go1.x
        - provided.al2
        - provided.al2023
      compatibleArchitectures:
        - x86_64
        - arm64
      licenseInfo: "BSD-3-Clause"