**Lambda Code Signing Config Data Source JSONC Example**

This example demonstrates how to retrieve a Lambda code signing configuration using the data source in JSONC format.

```javascript
{
  "variables": {
    "codeSigningConfigArn": {
      "type": "string",
      "description": "The ARN of the Lambda code signing configuration."
    }
  },
  "datasources": {
    "getCodeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "metadata": {
        "displayName": "Code Signing Configuration"
      },
      "filter": {
        "field": "arn",
        "operator": "=",
        "search": "${variables.codeSigningConfigArn}"
      },
      "exports": {
        "arn": {
          "type": "string",
          "aliasFor": "arn"
        },
        "codeSigningConfigId": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "allowedPublishers.signingProfileVersionArns": {
          "type": "array"
        },
        "codeSigningPolicies.untrustedArtifactOnDeployment": {
          "type": "string"
        }
      }
    }
  }
}
``` 