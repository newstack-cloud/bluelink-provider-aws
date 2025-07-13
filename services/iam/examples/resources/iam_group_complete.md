**IAM Group - Complete**

This example demonstrates creating an IAM group with inline policies and managed policies.

```yaml
resources:
  developers:
    type: aws/iam/group
    metadata:
      displayName: Developers Group
    spec:
      groupName: developers
      path: /
      # Inline policies
      policies:
        - policyName: EC2ReadOnly
          policyDocument:
            Version: "2012-10-17"
            Statement:
              - Effect: Allow
                Action:
                  - "ec2:Describe*"
                  - "ec2:Get*"
                Resource: "*"
      # Managed policies
      managedPolicyArns:
        - "arn:aws:iam::aws:policy/ReadOnlyAccess"
```

```javascript
{
  "resources": {
    "developers": {
      "type": "aws/iam/group",
      "metadata": {
        "displayName": "Developers Group"
      },
      "spec": {
        "groupName": "developers",
        "path": "/",
        // Inline policies
        "policies": [
          {
            "policyName": "EC2ReadOnly",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "ec2:Describe*",
                    "ec2:Get*"
                  ],
                  "Resource": "*"
                }
              ]
            }
          }
        ],
        // Managed policies
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/ReadOnlyAccess"
        ]
      }
    }
  }
}
``` 