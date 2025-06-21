The link type used to link a lambda function to another lambda function where the first function will be configured to be able to invoke and second function.

This will populate permissions and environment variables for the first function to be able to invoke the second function.

**Example for all target functions**

```yaml
resources:
    ordersFunction:
        type: aws/lambda/function
        metadata:
            displayName: Orders Function
            annotations:
                # Disable environment variable population which is enabled by default.
                aws.lambda.function.populateEnvVars: false
            labels:
                app: orders
        linkSelector:
            byLabel:
                app: orders
                system: global
        spec:
            handler: index.handler
            runtime: nodejs20.x
            code:
                s3Bucket: my-bucket
                s3Key: orders.zip

    logOrderEventsFunction:
        type: aws/lambda/function
        metadata:
            displayName: Log Order Events Function
            labels:
                app: orders
        spec:
            handler: index.handler

    logAllEventsFunction:
        type: aws/lambda/function
        metadata:
            displayName: Log All Events Function
            labels:
                system: global
        spec:
            handler: index.handler
```


**Example for a specific target function**

```yaml
resources:
    ordersFunction:
        type: aws/lambda/function
        metadata:
            displayName: Orders Function
            annotations:
                # These annotations will disable environment variable population for all
                # target functions except for the logOrderEventsFunction.
                aws.lambda.function.populateEnvVars: false
                aws.lambda.function.logOrderEventsFunction.populateEnvVars: true
                # This annotation will set the environment variable name for the logOrderEventsFunction reference in the ordersFunction.
                aws.lambda.function.logOrderEventsFunction.envVarName: AWS_LAMBDA_FUNCTION_LOG_ORDER_EVENTS
        linkSelector:
            byLabel:
                app: orders
                system: global
        spec:
            handler: index.handler
            runtime: nodejs20.x
            code:
                s3Bucket: my-bucket
                s3Key: orders.zip

    logOrderEventsFunction:
        type: aws/lambda/function
        metadata:
            displayName: Log Order Events Function
            labels:
                app: orders
        spec:
            handler: index.handler

    logAllEventsFunction:
        type: aws/lambda/function
        metadata:
            displayName: Log All Events Function
            labels:
                system: global
        spec:
            handler: index.handler
```
