The link type used to link a lambda function to a code signing config.
This will automatically populate the code signing config ARN of the lambda function
for any resources that match the link selector of the lambda function.

**Example**

```yaml
resources:
    ordersFunction:
        type: aws/lambda/function
        metadata:
            displayName: Orders Function
            labels:
                app: orders
        linkSelector:
            byLabel:
                app: orders
        spec:
            handler: index.handler
            runtime: nodejs20.x
            code:
                s3Bucket: my-bucket
                s3Key: orders.zip

    ordersCodeSigningConfig:
        type: aws/lambda/codeSigningConfig
        metadata:
            displayName: Orders Code Signing Config
            labels:
                app: orders
        spec:
            allowedPublishers:
                signingProfileVersionArns:
                    - arn:aws:signer:us-east-1:123456789012:signing-profile/orders-signing-profile
```
