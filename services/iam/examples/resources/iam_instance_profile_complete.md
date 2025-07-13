**IAM Instance Profile - Complete**

This example demonstrates creating an IAM instance profile with all available configuration options.

```yaml
resources:
  myInstanceProfile:
    type: aws/iam/instanceProfile
    metadata:
      displayName: My Instance Profile
    spec:
      instanceProfileName: MyInstanceProfile
      path: /
      role: MyRole
``` 