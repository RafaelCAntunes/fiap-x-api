package aws

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


type S3Adapter struct {
	Client *s3.Client
	Bucket string
}

func NewS3Adapter(cfg aws.Config, bucket string) *S3Adapter {
	return &S3Adapter{
		Client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.UsePathStyle = true
		}),
		Bucket: bucket,
	}
}

func (s *S3Adapter) UploadVideo(file io.Reader, fileName string) (string, error) {
	_, err := s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	return fileName, err
}

func (s *S3Adapter) EnsureBucketExists(ctx context.Context) error {
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.Bucket),
	})

	if err != nil {
		log.Printf("Bucket %s não encontrado. Tentando criar...", s.Bucket)
		_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(s.Bucket),
		})
		return err
	}
	return nil
}

func (s *S3Adapter) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}
