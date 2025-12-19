package aws

import (
	"context"
	"log"

	"checkin.service/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// NewAWSConfig creates a new AWS configuration, pointing to LocalStack if an endpoint is provided.
func NewAWSConfig(ctx context.Context, appConfig config.Config) (aws.Config, error) {
	if appConfig.IsLocalDev {
		log.Println("Local development mode detected. Routing AWS calls to LocalStack.")
		// This is the key part: We create a custom endpoint resolver.
		// If appConfig.AWSEndpoint is set, we tell the SDK to send all requests to that URL.
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if appConfig.AWSEndpoint != "" {
				return aws.Endpoint{
					URL:           appConfig.AWSEndpoint,
					SigningRegion: region,
					PartitionID:   "aws",
				}, nil
			}
			// Fallback to default AWS endpoint resolution if no custom endpoint is provided.
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		// Load the default config, but override the endpoint resolver and credentials for local development.
		return awsConfig.LoadDefaultConfig(ctx,
			awsConfig.WithRegion(appConfig.AWSRegion),
			awsConfig.WithEndpointResolverWithOptions(customResolver),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		)
	}

	// For non-local environments, use the standard AWS SDK config loading.
	// This will automatically use credentials from the environment (e.g., IAM role for service accounts).
	log.Println("Production mode detected. Using standard AWS credential chain.")
	return awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(appConfig.AWSRegion))
}
