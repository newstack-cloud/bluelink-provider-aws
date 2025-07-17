# IAM Server Certificate JSONC Example

```javascript
{
  "resources": {
    "jsoncCert": {
      "type": "aws/iam/serverCertificate",
      "metadata": {
        "displayName": "Jsonc Server Certificate",
        "description": "JSONC example of an IAM server certificate"
      },
      "spec": {
        "serverCertificateName": "JsoncCert",
        "certificateBody": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "privateKey": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
        "certificateChain": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "path": "/cloudfront/test/",
        "tags": [
          { "key": "Environment", "value": "Staging" },
          { "key": "Owner", "value": "DevOps" }
        ]
      }
    }
  }
}
``` 