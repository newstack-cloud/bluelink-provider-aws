# IAM Access Key with JSONC Comments

An IAM access key example with detailed comments explaining each field.

```javascript
{
  "resources": {
    "service_account_key": {
      "type": "aws/iam/accessKey",
      "metadata": {
        "displayName": "Service Account Access Key"
      },
      "spec": {
        // The username for which to create the access key
        "userName": "service-account",
        // Optional: Set the initial status of the access key
        // Valid values: "Active" or "Inactive"
        // Default: "Active"
        "status": "Active"
      }
    }
  }
}
``` 