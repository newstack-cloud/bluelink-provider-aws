**Lambda Function Version JSONC Example**

This example demonstrates how to create a Lambda function version using JSONC format.

```jsonc
{
  "resources": {
    "version1": {
      "type": "aws/lambda/functionVersion",
      "functionName": "my-lambda-function",
      "description": "Initial version of the function"  // Optional description
    }
  }
}
``` 