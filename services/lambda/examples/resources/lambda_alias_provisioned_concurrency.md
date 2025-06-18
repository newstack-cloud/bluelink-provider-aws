**Lambda Alias with Provisioned Concurrency**

This example demonstrates how to create a Lambda alias with provisioned concurrency for high-performance workloads.

```yaml
resources:
  highPerfAlias:
    type: aws/lambda/alias
    spec:
      functionName: my-lambda-function
      name: HIGHPERF
      functionVersion: "3"
      description: "High performance alias with provisioned concurrency"
      provisionedConcurrencyConfig:
        provisionedConcurrentExecutions: 10
``` 