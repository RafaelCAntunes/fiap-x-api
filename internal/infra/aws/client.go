package aws

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func NewAWSConfig(local bool) (aws.Config, error) {
	ctx := context.TODO()
	region := "us-east-1"
	

	if local {
		awsEndpoint := os.Getenv("AWS_ENDPOINT")
		if awsEndpoint == "" {
			awsEndpoint = "http://localhost:4566"
		}
		log.Printf("Ambiente LOCAL detectado. Endpoint: %s", awsEndpoint)

		return config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			// credenciais dummy para não falhar
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           awsEndpoint,
						SigningRegion: region,
					}, nil
				})),
		)
	}

	log.Println("Ambiente AWS detectado. Usando provedor de credenciais padrão.")
	return config.LoadDefaultConfig(ctx, config.WithRegion(region))
}
