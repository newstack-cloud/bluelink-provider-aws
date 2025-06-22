**IAM User - JSONC**

This example demonstrates creating an IAM user using JSONC format with comprehensive configuration.

```javascript
{
  "resources": {
    "developmentUser": {
      "type": "aws/iam/user",
      "metadata": {
        "displayName": "Development IAM User"
      },
      "spec": {
        "userName": "dev-api-user",
        "path": "/application/",
        // Add user to existing groups
        "groups": [
          "Developers",
          "APIUsers"
        ],
        // Attach AWS managed policies
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/ReadOnlyAccess"
        ],
        // Set permissions boundary
        "permissionsBoundary": "arn:aws:iam::aws:policy/PowerUserAccess",
        // Create console login profile
        "loginProfile": {
          "password": "${var.user_password}",
          "passwordResetRequired": true
        },
        // Define inline policies
        "policies": [
          {
            "policyName": "DynamoDBAccess",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "dynamodb:GetItem",
                    "dynamodb:PutItem",
                    "dynamodb:Query",
                    "dynamodb:Scan"
                  ],
                  "Resource": [
                    "arn:aws:dynamodb:us-east-1:123456789012:table/MyTable",
                    "arn:aws:dynamodb:us-east-1:123456789012:table/MyTable/index/*"
                  ]
                }
              ]
            }
          }
        ],
        // Resource tags
        "tags": [
          {
            "key": "Environment",
            "value": "Development"
          },
          {
            "key": "Application",
            "value": "MyApp"
          }
        ]
      }
    }
  }
}
```
