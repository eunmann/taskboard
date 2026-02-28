package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// DatabaseStack creates the RDS PostgreSQL instance.
type DatabaseStack struct {
	awscdk.Stack
	Instance awsrds.DatabaseInstance
}

// NewDatabaseStack creates the database infrastructure.
func NewDatabaseStack(
	scope constructs.Construct,
	id string,
	props *awscdk.StackProps,
	cfg config.EnvironmentConfig,
	network *NetworkStack,
) *DatabaseStack {
	stack := awscdk.NewStack(scope, &id, props)

	instance := awsrds.NewDatabaseInstance(stack, jsii.String("Database"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_16_4(),
		}),
		InstanceType:       awsec2.NewInstanceType(jsii.String(cfg.RdsInstanceClass)),
		Vpc:                network.VPC,
		VpcSubnets:         &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS},
		SecurityGroups:     &[]awsec2.ISecurityGroup{network.RDSSecurityGrp},
		DatabaseName:       jsii.String("platform"),
		AllocatedStorage:   jsii.Number(20),  //nolint:mnd // infrastructure constant
		MaxAllocatedStorage: jsii.Number(100), //nolint:mnd // infrastructure constant
		RemovalPolicy:      awscdk.RemovalPolicy_SNAPSHOT,
		DeletionProtection: jsii.Bool(true),
	})

	return &DatabaseStack{
		Stack:    stack,
		Instance: instance,
	}
}
