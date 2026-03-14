#!/bin/bash
echo "Configurando infraestrutura AWS Local..."

#Fila de Erro (DLQ)
awslocal sqs create-queue --queue-name video-queue-dlq

DLQ_ARN=$(awslocal sqs get-queue-attributes \
    --queue-url http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/video-queue-dlq \
    --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

# Fila Principal vinculada à DLQ (após 3 falhas mensagem vai pro DLQ)
awslocal sqs create-queue --queue-name video-queue \
    --attributes '{
        "RedrivePolicy": "{\"deadLetterTargetArn\":\"'"$DLQ_ARN"'\",\"maxReceiveCount\":\"3\"}",
        "VisibilityTimeout": "60"
    }'

awslocal s3api head-bucket --bucket fiap-x-videos-bucket || awslocal s3 mb s3://fiap-x-videos-bucket
echo "Infraestrutura local criada com sucesso!"