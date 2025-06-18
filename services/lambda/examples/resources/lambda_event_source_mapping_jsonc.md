**JSONC Event Source Mapping Example**

This example demonstrates how to create an event source mapping using JSONC format.

```javascript
{
  "resources": {
    "orderProcessorFunction": {
      "type": "aws/lambda/function",
      "spec": {
        "functionName": "order-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  console.log('Processing orders:', JSON.stringify(event, null, 2));\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "orderQueueMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "spec": {
        "functionName": "${resources.orderProcessorFunction.functionName}",
        "eventSourceArn": "arn:aws:sqs:us-east-1:123456789012:order-queue",
        "batchSize": 10,
        "enabled": true,
        "maximumBatchingWindowInSeconds": 5,
        "functionResponseTypes": ["ReportBatchItemFailures"],
        "filterCriteria": {
          "filters": [
            {
              "pattern": "{\"source\":[\"aws.sqs\"]}"
            }
          ]
        },
        "destinationConfig": {
          "onFailure": {
            "destination": "arn:aws:sqs:us-east-1:123456789012:dlq-queue"
          }
        },
        "tags": {
          "Environment": "Production",
          "Application": "OrderProcessing"
        }
      }
    }
  }
}
``` 