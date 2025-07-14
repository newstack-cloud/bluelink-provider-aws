# Basic IAM OIDC Provider

A basic IAM OIDC provider for GitHub Actions.

```yaml
resources:
  github_actions_oidc:
    type: aws/iam/oidcProvider
    metadata:
      displayName: GitHub Actions OIDC Provider
    spec:
      url: https://token.actions.githubusercontent.com
      clientIdList:
        - sts.amazonaws.com
      thumbprintList:
        - cf23df2207d99a74fbe169e3eba035e633b65d94
```