**DocumentDB Event Source Mapping**

This example demonstrates how to create an event source mapping for Amazon DocumentDB change streams.

```yaml
resources:
  documentdbProcessorFunction:
    type: aws/lambda/function
    spec:
      functionName: documentdb-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing DocumentDB change:', record);
            }
            return { statusCode: 200 };
          };

  documentdbMapping:
    type: aws/lambda/eventSourceMapping
    spec:
      functionName: ${resources.documentdbProcessorFunction.functionName}
      eventSourceArn: arn:aws:docdb:us-east-1:123456789012:cluster/my-cluster
      batchSize: 100
      enabled: true
      documentDbEventSourceConfig:
        databaseName: mydatabase
        collectionName: users
        fullDocument: UpdateLookup
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
``` 
