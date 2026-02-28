package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// StorageStack creates the S3 bucket.
type StorageStack struct {
	awscdk.Stack
	Bucket awss3.Bucket
}

// NewStorageStack creates the storage infrastructure.
func NewStorageStack(scope constructs.Construct, id string, props *awscdk.StackProps, cfg config.EnvironmentConfig) *StorageStack {
	stack := awscdk.NewStack(scope, &id, props)

	bucket := awss3.NewBucket(stack, jsii.String("Storage"), &awss3.BucketProps{
		BucketName:    jsii.String(cfg.BucketName("storage")),
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
		Versioned:     jsii.Bool(true),
		Encryption:    awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})

	return &StorageStack{
		Stack:  stack,
		Bucket: bucket,
	}
}
