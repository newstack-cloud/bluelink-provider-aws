**IAM User - Complete**

This example demonstrates creating an IAM user with all available configuration options.

```yaml
resources:
  completeUser:
    type: aws/iam/user
    metadata:
      displayName: Complete IAM User Configuration
    spec:
      userName: complete-user-example
      path: /developers/
      groups:
        - Developers
        - ReadOnlyUsers
      managedPolicyArns:
        - arn:aws:iam::aws:policy/ReadOnlyAccess
        - arn:aws:iam::123456789012:policy/MyCustomPolicy
      permissionsBoundary: arn:aws:iam::aws:policy/PowerUserAccess
      loginProfile:
        password: "MySecurePassword123!"
        passwordResetRequired: true
      policies:
        - policyName: S3Access
          policyDocument:
            Version: "2012-10-17"
            Statement:
              - Effect: Allow
                Action:
                  - s3:GetObject
                  - s3:PutObject
                Resource:
                  - arn:aws:s3:::my-bucket/*
        - policyName: EC2Describe
          policyDocument:
            Version: "2012-10-17"
            Statement:
              - Effect: Allow
                Action:
                  - ec2:DescribeInstances
                  - ec2:DescribeVolumes
                Resource: "*"
      tags:
        - key: Environment
          value: Development
        - key: Team
          value: Backend
        - key: Purpose
          value: API Access
```
