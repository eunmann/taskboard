package env

// PostgresHost and related constants define database connection env var keys.
const (
	PostgresHost     = "POSTGRES_HOST"
	PostgresPort     = "POSTGRES_PORT"
	PostgresUser     = "POSTGRES_USER"
	PostgresPassword = "POSTGRES_PASSWORD" //nolint:gosec // env var key name, not a credential
	PostgresDB       = "POSTGRES_DB"
	PostgresSSLMode  = "POSTGRES_SSLMODE"
)

// Port and related constants define application env var keys.
const (
	Port             = "PORT"
	SiteURL          = "SITE_URL"
	LocalAuthEnabled = "LOCAL_AUTH_ENABLED"
	CookieSecure     = "COOKIE_SECURE"
)

// AWSRegion and related constants define AWS env var keys.
const (
	AWSRegion      = "AWS_REGION"
	AWSAccessKeyID = "AWS_ACCESS_KEY_ID"
	AWSSecretKey   = "AWS_SECRET_ACCESS_KEY" //nolint:gosec // env var key name, not a credential
	AWSEndpointS3  = "AWS_ENDPOINT_URL_S3"
	AWSEndpointSQS = "AWS_ENDPOINT_URL_SQS"
	S3Bucket       = "S3_BUCKET"
)

// TaskQueueURL and related constants define SQS queue env var keys.
const (
	TaskQueueURL  = "TASK_QUEUE_URL"
	TaskDLQURL    = "TASK_DLQ_URL"
	EmailQueueURL = "EMAIL_QUEUE_URL"
)

// TaskEnabled and related constants define worker env var keys.
const (
	TaskEnabled      = "TASK_ENABLED"
	TaskConcurrency  = "TASK_CONCURRENCY"
	TaskPollInterval = "TASK_POLL_INTERVAL"
)

// FromEmail is the env var key for the sender email address.
const FromEmail = "FROM_EMAIL"

// LogLevel and LogFormat define logging env var keys.
const (
	LogLevel  = "LOG_LEVEL"
	LogFormat = "LOG_FORMAT"
)
