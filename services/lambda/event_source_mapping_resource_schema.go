package lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaEventSourceMappingResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaEventSourceMappingDefinition",
		Description: "The definition of an AWS Lambda event source mapping.",
		Required:    []string{"functionName"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			// Required fields
			"functionName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name or ARN of the Lambda function. Name formats: Function name (MyFunction), Function ARN (arn:aws:lambda:us-west-2:123456789012:function:MyFunction), Version or Alias ARN (arn:aws:lambda:us-west-2:123456789012:function:MyFunction:PROD), Partial ARN (123456789012:function:MyFunction).",
				MinLength:   1,
				MaxLength:   140,
				Pattern:     "(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}(-gov)?(-iso([a-z])?)?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:(\\$LATEST|[a-zA-Z0-9-_]+))?",
			},

			// Optional core configuration
			"eventSourceArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the event source. For Amazon Kinesis, the ARN of the data stream or stream consumer. For Amazon DynamoDB Streams, the ARN of the stream. For Amazon SQS, the ARN of the queue. For Amazon MSK, the ARN of the cluster. For Amazon MQ, the ARN of the broker. For Amazon DocumentDB, the ARN of the DocumentDB change stream.",
				Pattern:     "arn:(aws[a-zA-Z0-9-]*):([a-zA-Z0-9\\-])+:([a-z]{2}(-gov)?(-iso([a-z])?)?-[a-z]+-\\d{1})?(:\\d{12})?:(.*)",
				MinLength:   12,
				MaxLength:   1024,
			},
			"batchSize": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum number of records in each batch that Lambda pulls from your stream or queue and sends to your function. Default varies by service: Amazon SQS (10), all other services (100). Maximum: 10,000.",
				Minimum:     core.ScalarFromInt(1),
				Maximum:     core.ScalarFromInt(10000),
			},
			"enabled": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Whether the event source mapping is enabled.",
				Default:     core.MappingNodeFromBool(true),
			},
			"startingPosition": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The position in a stream from which to start reading. Required for Amazon Kinesis and Amazon DynamoDB Stream event sources.",
				Pattern:     "(LATEST|TRIM_HORIZON|AT_TIMESTAMP)+",
				MinLength:   6,
				MaxLength:   12,
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("TRIM_HORIZON"),
					core.MappingNodeFromString("LATEST"),
					core.MappingNodeFromString("AT_TIMESTAMP"),
				},
			},
			"startingPositionTimestamp": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "With StartingPosition set to AT_TIMESTAMP, the time from which to start reading, in Unix time seconds. Cannot be in the future.",
			},

			// Batching and retry configuration
			"maximumBatchingWindowInSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum amount of time, in seconds, that Lambda spends gathering records before invoking the function. 0-300 seconds.",
				Minimum:     core.ScalarFromInt(0),
				Maximum:     core.ScalarFromInt(300),
			},
			"maximumRecordAgeInSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "(Kinesis and DynamoDB Streams only) Discard records older than the specified age. The default value is -1 (infinite).",
				Minimum:     core.ScalarFromInt(-1),
				Maximum:     core.ScalarFromInt(604800),
			},
			"maximumRetryAttempts": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "(Kinesis and DynamoDB Streams only) Discard records after the specified number of retries. The default value is -1 (infinite).",
				Minimum:     core.ScalarFromInt(-1),
				Maximum:     core.ScalarFromInt(10000),
			},
			"bisectBatchOnFunctionError": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "(Kinesis and DynamoDB Streams only) If the function returns an error, split the batch in two and retry.",
			},
			"parallelizationFactor": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "(Kinesis and DynamoDB Streams only) The number of batches to process concurrently from each shard. Default: 1.",
				Default:     core.MappingNodeFromInt(1),
				Minimum:     core.ScalarFromInt(1),
				Maximum:     core.ScalarFromInt(10),
			},
			"tumblingWindowInSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "(Kinesis and DynamoDB Streams only) The duration in seconds of a processing window. A value of 0 indicates no tumbling window.",
				Minimum:     core.ScalarFromInt(0),
				Maximum:     core.ScalarFromInt(900),
			},

			// Function response configuration
			"functionResponseTypes": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "(Kinesis, DynamoDB Streams, and SQS) A list of current response type enums applied to the event source mapping.",
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
					AllowedValues: []*core.MappingNode{
						core.MappingNodeFromString("ReportBatchItemFailures"),
					},
				},
			},

			// Security and encryption
			"kmsKeyArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the AWS KMS customer managed key that Lambda uses to encrypt your function's filter criteria.",
				Pattern:     "(arn:(aws[a-zA-Z-]*)?:[a-z0-9-.]+:.*)|())",
				MinLength:   12,
				MaxLength:   2048,
			},

			// Filter criteria
			"filterCriteria": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "An object that defines the filter criteria that determine whether Lambda should process an event.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"filters": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "A list of filters.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeObject,
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"pattern": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "A filter pattern.",
								},
							},
						},
					},
				},
			},

			// Destination configuration
			"destinationConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "(Kinesis, DynamoDB Streams, Amazon MSK, and self-managed Kafka only) A configuration object that specifies the destination of an event after Lambda processes it.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"onFailure": {
						Type:        provider.ResourceDefinitionsSchemaTypeObject,
						Description: "The destination configuration for failed invocations.",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"destination": {
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The Amazon Resource Name (ARN) of the destination resource.",
							},
						},
					},
					"onSuccess": {
						Type:        provider.ResourceDefinitionsSchemaTypeObject,
						Description: "The destination configuration for successful invocations.",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"destination": {
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The Amazon Resource Name (ARN) of the destination resource.",
							},
						},
					},
				},
			},

			// Source access configurations
			"sourceAccessConfigurations": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "An array of authentication protocols, VPC components, or virtual host to secure and define your event source.",
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeObject,
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"type": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The type of authentication protocol, VPC component, or virtual host for your event source.",
						},
						"uri": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The value for your chosen configuration in Type.",
						},
					},
				},
			},

			// Amazon MSK configuration
			"amazonManagedKafkaEventSourceConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Specific configuration settings for an Amazon Managed Streaming for Apache Kafka (Amazon MSK) event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"consumerGroupId": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The identifier for the Kafka consumer group to join.",
					},
				},
			},

			// Self-managed Kafka configuration
			"selfManagedEventSource": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The self-managed Apache Kafka cluster for your event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"endpoints": {
						Type:        provider.ResourceDefinitionsSchemaTypeObject,
						Description: "The list of bootstrap servers for your Kafka brokers.",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"kafkaBootstrapServers": {
								Type:        provider.ResourceDefinitionsSchemaTypeArray,
								Description: "The list of bootstrap servers for your Kafka brokers.",
								Items: &provider.ResourceDefinitionsSchema{
									Type: provider.ResourceDefinitionsSchemaTypeString,
								},
							},
						},
					},
				},
			},

			"selfManagedKafkaEventSourceConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Specific configuration settings for a self-managed Apache Kafka event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"consumerGroupId": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The identifier for the Kafka consumer group to join.",
					},
				},
			},

			// DocumentDB configuration
			"documentDbEventSourceConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Specific configuration settings for a DocumentDB event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"databaseName": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The name of the database to consume within the DocumentDB cluster.",
					},
					"collectionName": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The name of the collection to consume within the database.",
					},
					"fullDocument": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "Determines what DocumentDB sends to your event stream when data changes.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("UpdateLookup"),
							core.MappingNodeFromString("Default"),
						},
					},
				},
			},

			// Metrics configuration
			"metricsConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The metrics configuration for your event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"metrics": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "A list of metrics to collect.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
					},
				},
			},

			// Scaling configurations
			"scalingConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "(Amazon SQS only) The scaling configuration for the event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"maximumConcurrency": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "Limits the number of concurrent instances that the Amazon SQS event source can invoke.",
					},
				},
			},

			"provisionedPollerConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "(Amazon MSK and self-managed Apache Kafka only) The provisioned mode configuration for the event source.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"minimumPollers": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The minimum number of pollers that are allocated to the event source.",
					},
					"maximumPollers": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The maximum number of pollers that are allocated to the event source.",
					},
				},
			},

			// MQ and Kafka specific
			"topics": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "The name of the Kafka topic.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:      provider.ResourceDefinitionsSchemaTypeString,
					MinLength: 1,
					MaxLength: 249,
					Pattern:   "^[^.]([a-zA-Z0-9\\-_.]+)",
				},
			},
			"queues": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "(MQ) The name of the Amazon MQ broker destination queue to consume.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:      provider.ResourceDefinitionsSchemaTypeString,
					MinLength: 1,
					MaxLength: 1000,
				},
			},

			// Tags
			"tags": lambdaSchemaTags("event source mapping"),

			// Computed fields returned by AWS
			"eventSourceMappingArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the event source mapping.",
				Computed:    true,
			},
			"functionArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the Lambda function.",
				Computed:    true,
			},
			"id": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The identifier of the event source mapping.",
				Computed:    true,
			},
			"state": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The state of the event source mapping.",
				Computed:    true,
			},
		},
	}
}
