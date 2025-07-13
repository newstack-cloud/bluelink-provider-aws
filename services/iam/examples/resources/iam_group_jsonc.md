**IAM Group - JSONC**

This example demonstrates creating an IAM group using JSONC format with comprehensive configuration.

```javascript
{
  "resources": {
    "adminGroup": {
      "type": "aws/iam/group",
      "metadata": {
        "displayName": "Administrators Group"
      },
      "spec": {
        // Group name (optional - will be auto-generated if not provided)
        "groupName": "admin-group",
        
        // Path for the group (defaults to "/" if not specified)
        "path": "/",
        
        // Inline policies attached to the group
        "policies": [
          {
            "policyName": "AdminPolicy",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": "*",
                  "Resource": "*"
                }
              ]
            }
          }
        ],
        
        // Managed policies to attach to the group
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/AdministratorAccess"
        ]
      }
    }
  }
}
``` 