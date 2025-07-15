# Complete IAM SAML Provider

This example demonstrates how to create a comprehensive IAM SAML provider with all available configuration options.

```yaml
resources:
  okta_saml:
    type: aws/iam/samlProvider
    metadata:
      displayName: Okta SAML Provider
      description: SAML provider for Okta SSO integration
      labels:
        provider: okta
        environment: production
    spec:
      name: OktaSAMLProvider
      samlMetadataDocument: |
        <?xml version="1.0" encoding="UTF-8"?>
        <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                          entityID="http://www.okta.com/exk1fxpisXtQDf6v4357">
          <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
            <KeyDescriptor use="signing">
              <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
                <X509Data>
                  <X509Certificate>MIIDpDCCAoygAwIBAgIGAVs7...</X509Certificate>
                </X509Data>
              </KeyInfo>
            </KeyDescriptor>
            <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>
            <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                               Location="https://company.okta.com/app/company_awsaccountid_1/exk1fxpisXtQDf6v4357/sso/saml"/>
            <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" 
                               Location="https://company.okta.com/app/company_awsaccountid_1/exk1fxpisXtQDf6v4357/sso/saml"/>
          </IDPSSODescriptor>
        </EntityDescriptor>
      tags:
        - key: Environment
          value: Production
        - key: Service
          value: SSO
        - key: Provider
          value: Okta
        - key: ManagedBy
          value: Security
```