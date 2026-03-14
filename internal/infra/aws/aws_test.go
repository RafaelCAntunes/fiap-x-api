package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestAWSConfig_Initialization(t *testing.T) {
	t.Run("Deve criar configuração local com sucesso", func(t *testing.T) {
		cfg, err := NewAWSConfig(true)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("Deve criar S3Adapter e SQSAdapter", func(t *testing.T) {
		cfg := aws.Config{}
		s3 := NewS3Adapter(cfg, "test-bucket")
		sqs := NewSQSAdapter(cfg, "http://sqs-url")

		assert.NotNil(t, s3)
		assert.Equal(t, "test-bucket", s3.Bucket)
		assert.NotNil(t, sqs)
		assert.Equal(t, "http://sqs-url", sqs.QueueURL)
	})
}
