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
        "assumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Service": [
                  "apigateway.amazonaws.com"
                ]
              },
              "Action": [
                "sts:AssumeRole"
              ]
            }
          ]
        },
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/service-role/AWSLambdaRole"
        ],
        "policies": [
          {
            "policyName": "CustomLogsPolicy",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "logs:CreateLogGroup",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                  ],
                  "Resource": [
                    "arn:aws:logs:*:*:*"
                  ]
                }
              ]
            }
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
