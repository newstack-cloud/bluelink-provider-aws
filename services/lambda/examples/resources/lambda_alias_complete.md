**Complete Lambda Alias Example**

This example demonstrates how to create a complete Lambda setup with a function, version, and alias.

```yaml
resources:
  myFunction:
    type: aws/lambda/function
    spec:
      functionName: my-sample-function
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            return {
              statusCode: 200,
              body: JSON.stringify('Hello from Lambda!')
            };
          };

  version1:
    type: aws/lambda/functionVersion
    spec:
      functionName: ${resources.myFunction.functionName}
      description: "Version 1"

  prodAlias:
    type: aws/lambda/alias
    spec:
      functionName: ${resources.myFunction.functionName}
      name: PROD
      functionVersion: ${resources.version1.version}
      description: "Production alias"
``` 