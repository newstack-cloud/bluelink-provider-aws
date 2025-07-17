# IAM Server Certificate Complete Example

```yaml
resources:
  cloudFrontCert:
    type: aws/iam/serverCertificate
    metadata:
      displayName: CloudFront Certificate
      description: Complete example of an IAM server certificate with multiple tags
    spec:
      serverCertificateName: CloudFrontCert
      certificateBody: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
      certificateChain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      path: /cloudfront/prod/
      tags:
        - key: Environment
          value: Production
        - key: Department
          value: Engineering
        - key: ManagedBy
          value: Automation
``` 