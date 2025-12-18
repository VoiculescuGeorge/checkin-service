#!/bin/bash

echo "---Creating SQS queues---"

awslocal sqs create-queue --queue-name labor-queue
awslocal sqs create-queue --queue-name email-queue

echo "---SQS queues created---"