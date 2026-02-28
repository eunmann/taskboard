// Package main is the CDK application entry point.
package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
	"github.com/OWNER/PROJECT_NAME/cdk/stacks"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Load environment configuration
	cdkEnv := getEnv("CDK_ENV", "prod")
	account := getEnv("CDK_DEFAULT_ACCOUNT", "")
	region := getEnv("CDK_DEFAULT_REGION", "us-east-1")

	var cfg config.EnvironmentConfig

	switch cdkEnv {
	case "prod":
		cfg = config.ProdConfig(account, region)
	default:
		cfg = config.ProdConfig(account, region)
		cfg.Name = cdkEnv
	}

	env := &awscdk.Environment{
		Account: jsii.String(cfg.Account),
		Region:  jsii.String(cfg.Region),
	}

	// Common tags
	awscdk.Tags_Of(app).Add(jsii.String("Environment"), jsii.String(cfg.Name), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Project"), jsii.String("App"), nil)
	awscdk.Tags_Of(app).Add(jsii.String("ManagedBy"), jsii.String("CDK"), nil)

	// ========================================================================
	// STATEFUL TIER (rarely changes)
	// ========================================================================

	networkStack := stacks.NewNetworkStack(app, cfg.StackName("Network"), &awscdk.StackProps{Env: env}, cfg)

	storageStack := stacks.NewStorageStack(app, cfg.StackName("Storage"), &awscdk.StackProps{Env: env}, cfg)

	queuesStack := stacks.NewQueuesStack(app, cfg.StackName("Queues"), &awscdk.StackProps{Env: env}, cfg)

	databaseStack := stacks.NewDatabaseStack(app, cfg.StackName("Database"), &awscdk.StackProps{Env: env}, cfg, networkStack)

	// ========================================================================
	// CI/CD
	// ========================================================================

	ecrStack := stacks.NewECRStack(app, cfg.StackName("ECR"), &awscdk.StackProps{Env: env}, cfg)

	// ========================================================================
	// COMPUTE TIER (frequent deploys)
	// ========================================================================

	ecsClusterStack := stacks.NewECSClusterStack(app, cfg.StackName("ECSCluster"), &awscdk.StackProps{Env: env}, cfg, networkStack)

	stacks.NewBackendServiceStack(
		app, cfg.StackName("BackendService"), &awscdk.StackProps{Env: env}, cfg,
		ecsClusterStack, databaseStack, storageStack, queuesStack, ecrStack,
	)

	// Suppress unused warnings for stacks that are dependencies only
	_ = networkStack
	_ = storageStack
	_ = queuesStack
	_ = databaseStack
	_ = ecrStack
	_ = ecsClusterStack

	app.Synth(nil)
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return defaultValue
}
