**JSONC Layer Version Data Source**

This example demonstrates how to define an AWS Lambda layer version data source.

```javascript
{
  "variables": {
    "layerArn": {
      "type": "string",
      "description": "The ARN of the layer version to retrieve."
    }
  },
  "datasources": {
    "externalLayer": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "External Library Layer"
      },
      "filter": [
        {
          "field": "layerName", 
          "operator": "=",
          "search": "${variables.layerArn}"
        },
        {
          "field": "versionNumber",
          "operator": "=", 
          "search": 2
        }
      ],
      "exports": {
        "arn": {
          "type": "string"
        },
        "layerVersionArn": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "compatibleRuntimes": {
          "type": "array"
        },
        "compatibleArchitectures": {
          "type": "array"
        }
      }
    }
  }
}```
