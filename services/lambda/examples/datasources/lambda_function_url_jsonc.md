**JSONC Function URL Data Source**

This example demonstrates how to define an AWS Lambda function URL data source.

```javascript
{
  "variables": {
    "functionName": {
      "type": "string",
      "description": "The name of the Lambda function to retrieve the function URL for."
    }
  },
  "datasources": {
    "getFunctionUrl": {
      "type": "aws/lambda/function_url",
      "metadata": {
        "displayName": "Lambda Function URL"
      },
      "filter": {
        "field": "functionName",
        "operator": "=",
        "search": "${variables.functionName}"
      },
      "exports": {
        "functionUrl": {
          "type": "string",
          "aliasFor": "functionUrl"
        },
        "authType": {
          "type": "string"
        },
        "cors": {
          "type": "object"
        }
      }
    }
  }
}
```
