# Lambda Code Signing Config Complete Example

```yaml
resources:
  codeSigningConfig:
    type: aws/lambda/codeSigningConfig
    spec:
      allowedPublishers:
        signingProfileVersionArns:
          - arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12
          - arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile2/ghijkl34
      codeSigningPolicies:
        untrustedArtifactOnDeployment: Enforce
      description: "Production code signing configuration"
      tags:
        - key: Environment
          value: Production
        - key: Team
          value: Backend
        - key: Project
          value: MyApplication