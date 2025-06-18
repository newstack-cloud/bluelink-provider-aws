**YAML Function Data Source - Export All Fields**

This example demonstrates how to define an AWS Lambda function data source
and export all fields to be used in other elements of the blueprint.

```yaml
variables:
  orderFunctionArn:
    type: string
    description: The ARN of the order retrieval function managed externally.

datasources:
  getOrderFunction:
	type: aws/lambda/function
	metadata:
	  displayName: Order Retrieval Function
	filter:
      field: arn
      operator: "="
      search: ${variables.orderFunctionArn}
  # Export all fields to be used in other elements of the blueprint.
  exports: "*"
```
