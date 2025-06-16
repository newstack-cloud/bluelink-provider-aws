# Lambda Code Signing Config JSONC Example

```javascript
{
  "resources": {
    "codeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "spec": {
        "allowedPublishers": {
          "signingProfileVersionArns": [
            "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12"
          ]
        },
        "codeSigningPolicies": {
          "untrustedArtifactOnDeployment": "Warn"
        },
        "description": "Development code signing configuration",
        "tags": [
          {
            "key": "Environment",
            "value": "Development"
          }
        ]
      }
    }
  }
}