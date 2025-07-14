# IAM OIDC Provider with JSONC Comments

An IAM OIDC provider example with detailed comments explaining each field.

```javascript
{
  "resources": {
    "gitlab_oidc": {
      "type": "aws/iam/oidcProvider",
      "metadata": {
        "displayName": "GitLab OIDC Provider",
        "description": "OIDC provider for GitLab CI/CD"
      },
      "spec": {
        // The URL of the identity provider.
        // The URL must begin with https:// and should correspond to the iss claim
        // in the provider's OpenID Connect ID tokens.
        "url": "https://gitlab.com",
        
        // A list of client IDs (also known as audiences) for the IAM OIDC provider.
        // When a mobile or web app registers with an OpenID Connect provider,
        // they establish a value that identifies the application.
        "clientIdList": [
          "https://gitlab.com"
        ],
        
        // A list of server certificate thumbprints for the OpenID Connect (OIDC)
        // identity provider's server certificates. Typically this list includes
        // only one entry. However, IAM lets you have up to five thumbprints
        // for an OIDC provider. This lets you maintain multiple thumbprints
        // if the identity provider is rotating certificates.
        "thumbprintList": [
          "b3dd7606d2b5a8b4a13771dbecc9ee1cecafa38a"
        ],
        
        // Optional: Tags to attach to the OIDC provider
        "tags": [
          {
            "key": "Environment",
            "value": "Development"
          },
          {
            "key": "CI/CD",
            "value": "GitLab"
          }
        ]
      }
    }
  }
}
```