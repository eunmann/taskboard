package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"github.com/OWNER/PROJECT_NAME/cdk/config"
)

// QueuesStack creates the SQS queues.
type QueuesStack struct {
	awscdk.Stack
	TaskQueue  awssqs.Queue
	TaskDLQ    awssqs.Queue
	EmailQueue awssqs.Queue
}

// NewQueuesStack creates the queue infrastructure.
func NewQueuesStack(scope constructs.Construct, id string, props *awscdk.StackProps, cfg config.EnvironmentConfig) *QueuesStack {
	stack := awscdk.NewStack(scope, &id, props)

	taskDLQ := awssqs.NewQueue(stack, jsii.String("TaskDLQ"), &awssqs.QueueProps{
		QueueName:         jsii.String(cfg.QueueName("task-dlq")),
		RetentionPeriod:   awscdk.Duration_Days(jsii.Number(14)), //nolint:mnd // infrastructure constant
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(30)), //nolint:mnd // infrastructure constant
	})

	taskQueue := awssqs.NewQueue(stack, jsii.String("TaskQueue"), &awssqs.QueueProps{
		QueueName:         jsii.String(cfg.QueueName("task-queue")),
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(120)), //nolint:mnd // infrastructure constant
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			Queue:           taskDLQ,
			MaxReceiveCount: jsii.Number(3), //nolint:mnd // infrastructure constant
		},
	})

	emailQueue := awssqs.NewQueue(stack, jsii.String("EmailQueue"), &awssqs.QueueProps{
		QueueName:         jsii.String(cfg.QueueName("email-queue")),
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(60)), //nolint:mnd // infrastructure constant
	})

	return &QueuesStack{
		Stack:      stack,
		TaskQueue:  taskQueue,
		TaskDLQ:    taskDLQ,
		EmailQueue: emailQueue,
	}
}
