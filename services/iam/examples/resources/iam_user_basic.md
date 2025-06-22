**IAM User - Basic**

This example demonstrates creating a basic IAM user with minimal configuration.

```yaml
resources:
  myUser:
    type: aws/iam/user
    metadata:
      displayName: My Basic IAM User
    spec:
      userName: my-basic-user
      path: /
```
