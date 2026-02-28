package config

// ProdConfig returns the production environment configuration.
func ProdConfig(account, region string) EnvironmentConfig {
	return EnvironmentConfig{
		Name:             "prod",
		Account:          account,
		Region:           region,
		VpcCidr:          "10.2.0.0/16", // TODO: adjust CIDR for your VPC
		EcsDesiredCount:  2,              //nolint:mnd // infrastructure constant
		UseFargateSpot:   true,
		RdsInstanceClass: "db.t3.micro", // TODO: adjust instance class for production
		DomainName:       "",            // TODO: set your domain name
		SiteURL:          "",            // TODO: set your site URL
	}
}
