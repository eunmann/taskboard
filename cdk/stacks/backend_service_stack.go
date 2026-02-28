package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// BackendServiceStack creates the ECS backend service.
type BackendServiceStack struct {
	awscdk.Stack
}

// NewBackendServiceStack creates the backend ECS service.
func NewBackendServiceStack(
	scope constructs.Construct,
	id string,
	props *awscdk.StackProps,
	cfg config.EnvironmentConfig,
	ecsCluster *ECSClusterStack,
	database *DatabaseStack,
	storage *StorageStack,
	queues *QueuesStack,
	ecr *ECRStack,
) *BackendServiceStack {
	stack := awscdk.NewStack(scope, &id, props)

	taskDef := awsecs.NewFargateTaskDefinition(stack, jsii.String("BackendTask"), &awsecs.FargateTaskDefinitionProps{
		MemoryLimitMiB: jsii.Number(512),  //nolint:mnd // infrastructure constant
		Cpu:            jsii.Number(256),   //nolint:mnd // infrastructure constant
		Family:         jsii.String(cfg.ResourceName("backend")),
	})

	taskDef.AddContainer(jsii.String("Backend"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromEcrRepository(ecr.BackendRepo, jsii.String("latest")),
		Logging: awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("backend"),
			LogRetention: awslogs.RetentionDays_TWO_WEEKS,
		}),
		PortMappings: &[]*awsecs.PortMapping{
			{ContainerPort: jsii.Number(8080)}, //nolint:mnd // app port
		},
		HealthCheck: &awsecs.HealthCheck{
			Command: &[]*string{
				jsii.String("CMD-SHELL"),
				jsii.String("wget -q -O /dev/null http://localhost:8080/health || exit 1"),
			},
		},
	})

	service := awsecs.NewFargateService(stack, jsii.String("BackendService"), &awsecs.FargateServiceProps{
		Cluster:        ecsCluster.Cluster,
		TaskDefinition: taskDef,
		DesiredCount:   jsii.Number(float64(cfg.EcsDesiredCount)),
		ServiceName:    jsii.String(cfg.ResourceName("backend")),
	})

	// Register with ALB target group
	targetGroup := ecsCluster.Listener.AddTargets(jsii.String("BackendTG"), &awselasticloadbalancingv2.AddApplicationTargetsProps{
		Port:    jsii.Number(8080), //nolint:mnd // app port
		Targets: &[]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{service},
		HealthCheck: &awselasticloadbalancingv2.HealthCheck{
			Path:            jsii.String("/health"),
			HealthyHttpCodes: jsii.String("200"),
		},
	})

	// Grant access to storage and queues
	storage.Bucket.GrantReadWrite(taskDef.TaskRole(), nil)
	queues.TaskQueue.GrantSendMessages(taskDef.TaskRole())
	queues.EmailQueue.GrantSendMessages(taskDef.TaskRole())

	// Suppress unused
	_ = database
	_ = targetGroup

	return &BackendServiceStack{Stack: stack}
}
