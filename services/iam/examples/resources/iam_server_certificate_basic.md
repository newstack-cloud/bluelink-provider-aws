# IAM Server Certificate Basic Example

```yaml
resources:
  myServerCertificate:
    type: aws/iam/serverCertificate
    metadata:
      displayName: My Server Certificate
      description: Basic example of an IAM server certificate
    spec:
      serverCertificateName: MyServerCertificate
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
      path: /
      tags:
        - key: Environment
          value: Production
``` 