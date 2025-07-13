**IAM Group - Basic**

This example demonstrates creating a basic IAM group with minimal configuration.

```yaml
resources:
  developers:
    type: aws/iam/group
    metadata:
      displayName: Developers Group
    spec:
      groupName: developers
      path: /
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
        "path": "/"
      }
    }
  }
}
``` 