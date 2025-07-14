# IAM Managed Policy (JSONC)

This example shows how to create an IAM managed policy using JSONC format with comments.

```jsonc
{
  // Policy name - must be unique within the account
  "policyName": "LambdaExecutionPolicy",
  
  // Policy document in JSON format
  "policyDocument": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "lambda:InvokeFunction",
          "lambda:GetFunction",
          "lambda:ListFunctions"
        ],
        "Resource": [
          // Allow access to specific Lambda functions
          "arn:aws:lambda:us-east-1:123456789012:function:my-function",
          "arn:aws:lambda:us-east-1:123456789012:function:my-other-function"
        ]
      },
      {
        "Effect": "Allow",
        "Action": [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        "Resource": [
          // Allow CloudWatch Logs access for Lambda execution
          "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/*"
        ]
      }
    ]
  },
  
  // Optional description
  "description": "Policy for Lambda function execution with CloudWatch Logs access",
  
  // Optional path - defaults to "/" if not specified
  "path": "/lambda-policies/",
  
  // Optional tags for resource management
  "tags": [
    {
      "key": "Service",
      "value": "Lambda"
    },
    {
      "key": "Environment",
      "value": "Development"
    },
    {
      "key": "Team",
      "value": "Backend"
    }
  ]
}
``` 