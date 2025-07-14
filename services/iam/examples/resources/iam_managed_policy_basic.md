# Basic IAM Managed Policy

This example creates a simple IAM managed policy that allows read-only access to S3 buckets.

```javascript
{
  "policyName": "S3ReadOnlyPolicy",
  "policyDocument": {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "s3:GetObject",
          "s3:ListBucket"
        ],
        "Resource": [
          "arn:aws:s3:::my-bucket",
          "arn:aws:s3:::my-bucket/*"
        ]
      }
    ]
  },
  "description": "Policy that allows read-only access to S3 bucket",
  "path": "/",
  "tags": [
    {
      "key": "Environment",
      "value": "Production"
    },
    {
      "key": "Project",
      "value": "MyProject"
    }
  ]
}
``` 