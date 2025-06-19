**Basic Lambda Alias Data Source**

This example demonstrates how to retrieve a Lambda alias using the data source.

```yaml
variables:
  functionName:
    type: string
    description: The name of the Lambda function.
  aliasName:
    type: string
    description: The name of the alias to retrieve.

datasources:
  getProductionAlias:
    type: aws/lambda/alias
    metadata:
      displayName: Production Alias
    filter:
      - field: functionName
        operator: "="
        search: ${variables.functionName}
      - field: name
        operator: "="
        search: ${variables.aliasName}
    exports:
      name:
        type: string
        aliasFor: name
      functionVersion:
        type: string
      description:
        type: string
      invokeArn:
        type: string
``` 