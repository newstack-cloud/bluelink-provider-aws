**Basic Layer Version Data Source**

This example demonstrates how to define an AWS Lambda layer version data source.

```yaml
variables:
  layerName:
    type: string
    description: The name of the layer to retrieve
    default: my-python-utils
  layerVersion:
    type: integer  
    description: The version number of the layer
    default: 1

datasources:
  pythonUtilsLayer:
    type: aws/lambda/layerVersion
    metadata:
      displayName: Python Utils Layer
    filter:
      - field: layerName
        operator: "="
        search: ${variables.layerName}
      - field: versionNumber
        operator: "="
        search: ${variables.layerVersion}
    exports:
      arn:
        type: string
        aliasFor: arn
      layerVersionArn:
        type: string
      compatibleRuntimes:
        type: array
      version:
        type: integer
```
