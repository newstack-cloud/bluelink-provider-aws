# Basic IAM SAML Provider

A basic IAM SAML provider for corporate SSO.

```yaml
resources:
  corporate_saml:
    type: aws/iam/samlProvider
    metadata:
      displayName: Corporate SAML Provider
    spec:
      name: CorporateSAML
      samlMetadataDocument: |
        <?xml version="1.0"?>
        <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                          entityID="http://www.example.com/saml">
          <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
            <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                               Location="https://www.example.com/saml/sso"/>
          </IDPSSODescriptor>
        </EntityDescriptor>
```