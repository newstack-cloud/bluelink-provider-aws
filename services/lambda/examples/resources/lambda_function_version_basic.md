**Basic Lambda Function Version**

This example demonstrates how to create a basic Lambda function version.

```yaml
resources:
  version1:
    type: aws/lambda/functionVersion
    functionName: my-lambda-function
    description: "Initial version of the function"
``` 