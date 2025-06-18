**JSONC Function Data Source**

This example demonstrates how to define an AWS Lambda function data source.

```javascript
{
  "variables": {
    "orderFunctionArn": {
      "type": "string",
      "description": "The ARN of the order retrieval function managed externally."
    }
  },
  "datasources": {
    "getOrderFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Order Retrieval Function"
      },
      "filter": {
        "field": "arn",
        "operator": "=",
        "search": "${variables.orderFunctionArn}"
      },
      "exports": {
        "name": {
          "type": "string",
          "aliasFor": "name"
        },
        "qualifiedArn": {
          "type": "string"
        },
        "layers": {
          "type": "array"
        }
      }
    }
  }
}
```
