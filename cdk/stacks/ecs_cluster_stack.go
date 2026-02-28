package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// ECSClusterStack creates the ECS cluster and ALB.
type ECSClusterStack struct {
	awscdk.Stack
	Cluster      awsecs.Cluster
	LoadBalancer awselasticloadbalancingv2.ApplicationLoadBalancer
	Listener     awselasticloadbalancingv2.ApplicationListener
}

// NewECSClusterStack creates the compute cluster infrastructure.
func NewECSClusterStack(
	scope constructs.Construct,
	id string,
	props *awscdk.StackProps,
	cfg config.EnvironmentConfig,
	network *NetworkStack,
) *ECSClusterStack {
	stack := awscdk.NewStack(scope, &id, props)

	cluster := awsecs.NewCluster(stack, jsii.String("Cluster"), &awsecs.ClusterProps{
		Vpc:         network.VPC,
		ClusterName: jsii.String(cfg.ResourceName("cluster")),
	})

	alb := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, jsii.String("ALB"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:              network.VPC,
		InternetFacing:   jsii.Bool(true),
		SecurityGroup:    network.ALBSecurityGrp,
		LoadBalancerName: jsii.String(cfg.ResourceName("alb")),
	})

	listener := alb.AddListener(jsii.String("HTTPListener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
		Port: jsii.Number(80), //nolint:mnd // HTTP port
	})

	// Default action: fixed 404
	listener.AddAction(jsii.String("Default"), &awselasticloadbalancingv2.AddApplicationActionProps{
		Action: awselasticloadbalancingv2.ListenerAction_FixedResponse(jsii.Number(404), //nolint:mnd // HTTP status
			&awselasticloadbalancingv2.FixedResponseOptions{
				ContentType: jsii.String("text/plain"),
				MessageBody: jsii.String("Not Found"),
			}),
	})

	return &ECSClusterStack{
		Stack:        stack,
		Cluster:      cluster,
		LoadBalancer: alb,
		Listener:     listener,
	}
}
