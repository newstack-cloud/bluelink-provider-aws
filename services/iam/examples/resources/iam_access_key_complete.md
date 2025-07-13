# Complete IAM Access Key

A complete IAM access key with all available options.

```yaml
resources:
  admin_access_key:
    type: aws/iam/accessKey
    metadata:
      displayName: Admin Access Key
    spec:
      userName: admin.user
      status: Active
``` 