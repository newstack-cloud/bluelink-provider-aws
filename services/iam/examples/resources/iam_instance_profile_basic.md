**IAM Instance Profile - Basic**

This example demonstrates creating a basic IAM instance profile with minimal configuration.

```yaml
resources:
  myInstanceProfile:
    type: aws/iam/instanceProfile
    metadata:
      displayName: My Instance Profile
    spec:
      role: MyRole
``` 