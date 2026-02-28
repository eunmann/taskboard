// Package stacks contains CDK stack definitions.
package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// NetworkStack creates the VPC and security groups.
type NetworkStack struct {
	awscdk.Stack
	VPC            awsec2.Vpc
	ALBSecurityGrp awsec2.SecurityGroup
	ECSSecurityGrp awsec2.SecurityGroup
	RDSSecurityGrp awsec2.SecurityGroup
}

// NewNetworkStack creates the network infrastructure.
func NewNetworkStack(scope constructs.Construct, id string, props *awscdk.StackProps, cfg config.EnvironmentConfig) *NetworkStack {
	stack := awscdk.NewStack(scope, &id, props)

	vpc := awsec2.NewVpc(stack, jsii.String("VPC"), &awsec2.VpcProps{
		VpcName:            jsii.String(cfg.ResourceName("vpc")),
		IpAddresses:        awsec2.IpAddresses_Cidr(jsii.String(cfg.VpcCidr)),
		MaxAzs:             jsii.Number(2), //nolint:mnd // infrastructure constant
		NatGateways:        jsii.Number(0),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
	})

	albSG := awsec2.NewSecurityGroup(stack, jsii.String("ALBSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(cfg.ResourceName("alb-sg")),
		Description:       jsii.String("ALB security group"),
		AllowAllOutbound:  jsii.Bool(true),
	})

	ecsSG := awsec2.NewSecurityGroup(stack, jsii.String("ECSSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(cfg.ResourceName("ecs-sg")),
		Description:       jsii.String("ECS tasks security group"),
		AllowAllOutbound:  jsii.Bool(true),
	})

	rdsSG := awsec2.NewSecurityGroup(stack, jsii.String("RDSSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(cfg.ResourceName("rds-sg")),
		Description:       jsii.String("RDS security group"),
		AllowAllOutbound:  jsii.Bool(false),
	})

	// Allow ALB -> ECS on port 8080
	ecsSG.AddIngressRule(albSG, awsec2.Port_Tcp(jsii.Number(8080)), jsii.String("ALB to ECS"), nil) //nolint:mnd // port number

	// Allow ECS -> RDS on port 5432
	rdsSG.AddIngressRule(ecsSG, awsec2.Port_Tcp(jsii.Number(5432)), jsii.String("ECS to RDS"), nil) //nolint:mnd // port number

	return &NetworkStack{
		Stack:          stack,
		VPC:            vpc,
		ALBSecurityGrp: albSG,
		ECSSecurityGrp: ecsSG,
		RDSSecurityGrp: rdsSG,
	}
}
