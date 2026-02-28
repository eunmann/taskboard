// Package config holds CDK stack configuration per environment.
package config

import "fmt"

// EnvironmentConfig defines infrastructure parameters per environment.
type EnvironmentConfig struct {
	Name             string
	Account          string
	Region           string
	VpcCidr          string
	EcsDesiredCount  int
	UseFargateSpot   bool
	RdsInstanceClass string
	DomainName       string
	SiteURL          string
}

// StackName returns a prefixed stack name.
func (c EnvironmentConfig) StackName(component string) string {
	return fmt.Sprintf("App-%s-%s", c.Name, component)
}

// ResourceName returns a prefixed resource name.
func (c EnvironmentConfig) ResourceName(resource string) string {
	return fmt.Sprintf("app-%s-%s", c.Name, resource)
}

// BucketName returns an S3 bucket name.
func (c EnvironmentConfig) BucketName(suffix string) string {
	return fmt.Sprintf("app-%s-%s-%s", c.Name, c.Account, suffix)
}

// QueueName returns an SQS queue name.
func (c EnvironmentConfig) QueueName(name string) string {
	return fmt.Sprintf("app-%s-%s", c.Name, name)
}
