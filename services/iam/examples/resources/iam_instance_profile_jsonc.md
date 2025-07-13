**IAM Instance Profile - JSONC**

This example demonstrates creating an IAM instance profile using JSONC format.

```javascript
{
  "resources": {
    "myInstanceProfile": {
      "type": "aws/iam/instanceProfile",
      "metadata": {
        "displayName": "My Instance Profile"
      },
      "spec": {
        "instanceProfileName": "MyInstanceProfile",
        "path": "/",
        "role": "MyRole"
      }
    }
  }
}
``` 