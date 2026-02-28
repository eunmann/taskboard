package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// ECRStack creates ECR repositories.
type ECRStack struct {
	awscdk.Stack
	BackendRepo awsecr.Repository
}

// NewECRStack creates the container registry infrastructure.
func NewECRStack(scope constructs.Construct, id string, props *awscdk.StackProps, cfg config.EnvironmentConfig) *ECRStack {
	stack := awscdk.NewStack(scope, &id, props)

	backendRepo := awsecr.NewRepository(stack, jsii.String("BackendRepo"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String(cfg.ResourceName("backend")),
		RemovalPolicy:  awscdk.RemovalPolicy_RETAIN,
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				MaxImageCount: jsii.Number(10), //nolint:mnd // infrastructure constant
				Description:   jsii.String("Keep last 10 images"),
			},
		},
	})

	return &ECRStack{
		Stack:       stack,
		BackendRepo: backendRepo,
	}
}
