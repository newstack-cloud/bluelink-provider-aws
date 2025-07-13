# Basic IAM Access Key

A basic IAM access key for a user.

```yaml
resources:
  john_access_key:
    type: aws/iam/accessKey
    metadata:
      displayName: John's Access Key
    spec:
      userName: john.doe
``` 