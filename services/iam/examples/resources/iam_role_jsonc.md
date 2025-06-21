**JSONC**

```javascript
{
  "resources": {
    "apiGatewayRole": {
      "type": "aws/iam/role",
      "metadata": {
        "displayName": "API Gateway Execution Role",
        "description": "IAM role for API Gateway to invoke Lambda functions",
        "labels": {
          "app": "api-gateway",
          "tier": "integration"
        }
      },
      "spec": {
        "roleName": "api-gateway-lambda-role",
        "description": "Role allowing API Gateway to invoke Lambda functions",
        "assumeRolePolicyDocument": "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Principal\": {\n        \"Service\": \"apigateway.amazonaws.com\"\n      },\n      \"Action\": \"sts:AssumeRole\"\n    }\n  ]\n}",
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/service-role/AWSLambdaRole"
        ],
        "policies": [
          {
            "policyName": "CustomLogsPolicy",
            "policyDocument": "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"logs:CreateLogGroup\",\n        \"logs:CreateLogStream\",\n        \"logs:PutLogEvents\"\n      ],\n      \"Resource\": \"arn:aws:logs:*:*:*\"\n    }\n  ]\n}"
          }
        ],
        "tags": [
          {
            "key": "Environment",
            "value": "Development"
          },
          {
            "key": "Service",
            "value": "API Gateway"
          }
        ]
      }
    }
  }
}
```
