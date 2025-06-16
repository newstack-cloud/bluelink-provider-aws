**Lambda Alias JSONC Example**

This example demonstrates how to create a Lambda alias using JSONC format.

```javascript
{
  "resources": {
    "productionAlias": {
      "type": "aws/lambda/alias",
      "functionName": "my-lambda-function",
      "name": "PROD",
      "functionVersion": "1",
      "description": "Production alias for my Lambda function",
      "routingConfig": {
        "additionalVersionWeights": {
          "2": 0.1  // Route 10% traffic to version 2
        }
      },
      "provisionedConcurrencyConfig": {
        "provisionedConcurrentExecutions": 10
      }
    }
  }
}
``` 