package aws

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSAdapter struct {
	Client   *sqs.Client
	QueueURL string
}

type VideoMessage struct {
	VideoID string `json:"video_id"`
}

func NewSQSAdapter(cfg aws.Config, queueURL string) *SQSAdapter {
	return &SQSAdapter{
		Client:   sqs.NewFromConfig(cfg),
		QueueURL: queueURL,
	}
}

func (s *SQSAdapter) PublishVideoJob(videoID string) error {
	msg := VideoMessage{VideoID: videoID}
	body, _ := json.Marshal(msg)

	_, err := s.Client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.QueueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}