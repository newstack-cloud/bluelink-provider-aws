**Lambda Alias with Traffic Routing**

This example demonstrates how to create a Lambda alias with traffic routing configuration for canary deployments.

```yaml
resources:
  canaryAlias:
    type: aws/lambda/alias
    spec:
      functionName: my-lambda-function
      name: CANARY
      functionVersion: "2"
      description: "Canary deployment with traffic splitting"
      routingConfig:
        additionalVersionWeights:
          "1": 0.1  # Route 10% traffic to version 1
``` 