# Lambda Code Signing Config Basic Example

```yaml
resources:
  codeSigningConfig:
    type: aws/lambda/codeSigningConfig
    spec:
      allowedPublishers:
        signingProfileVersionArns:
          - arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12