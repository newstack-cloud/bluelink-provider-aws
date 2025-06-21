**YAML Basic IAM Role**

This example demonstrates how to create a basic IAM role with a single service principal.

```yaml
resources:
  lambdaExecutionRole:
    type: aws/iam/role
    metadata:
      displayName: Lambda Execution Role
      description: IAM role for Lambda function execution
      labels:
        app: lambda
    spec:
      roleName: lambda-execution-role
      assumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Principal:
              Service:
                - lambda.amazonaws.com
            Action:
              - sts:AssumeRole
      managedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
      tags:
        Environment: Production
        Service: Lambda
```
