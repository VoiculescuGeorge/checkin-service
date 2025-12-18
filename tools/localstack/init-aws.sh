#!/bin/bash

echo "---"
echo "Initializing LocalStack AWS resources"
echo "---"

awslocal sqs create-queue --queue-name labor-queue-dlq
awslocal sqs create-queue --queue-name email-queue-dlq

LABOR_DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url http://localhost:4566/000000000000/labor-queue-dlq --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)
EMAIL_DLQ_ARN=$(awslocal sqs get-queue-attributes --queue-url http://localhost:4566/000000000000/email-queue-dlq --attribute-names QueueArn --query 'Attributes.QueueArn' --output text)

echo "Labor DLQ ARN: $LABOR_DLQ_ARN"
echo "Email DLQ ARN: $EMAIL_DLQ_ARN"

# Create the main labor queue with a redrive policy
awslocal sqs create-queue --queue-name labor-queue \
    --attributes '{
        "RedrivePolicy": "{\"deadLetterTargetArn\":\"'"$LABOR_DLQ_ARN"'\",\"maxReceiveCount\":\"5\"}"
    }'

# Create the main email queue with a redrive policy
awslocal sqs create-queue --queue-name email-queue \
    --attributes '{
        "RedrivePolicy": "{\"deadLetterTargetArn\":\"'"$EMAIL_DLQ_ARN"'\",\"maxReceiveCount\":\"5\"}"
    }'

echo "SQS Queues and DLQs created and configured."