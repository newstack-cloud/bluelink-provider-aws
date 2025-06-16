**Basic Lambda Alias**

This example demonstrates how to create a basic Lambda alias that points to a specific function version.

```yaml
resources:
  productionAlias:
    type: aws/lambda/alias
    functionName: my-lambda-function
    name: PROD
    functionVersion: "1"
    description: "Production alias for my Lambda function"
``` 