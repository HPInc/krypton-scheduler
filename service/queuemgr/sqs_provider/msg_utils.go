package sqs_provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hpinc/krypton-scheduler/service/db"
	"go.uber.org/zap"
)

// Receive a single message from the scheduler queue corresponding to the
// specified queue URL.
func (p *SqsQueueProvider) receiveMessage(queueUrl string) (*sqs.ReceiveMessageOutput,
	error) {
	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	msgResult, err := p.gSQS.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: 1,
		VisibilityTimeout:   awsSqsVisibilityTimeout,
		WaitTimeSeconds:     p.queueConfig.WatchDelay,
	})
	if err != nil {
		return nil, err
	}

	if len(msgResult.Messages) == 0 {
		return nil, nil
	}

	return msgResult, nil
}

// Send a single message to the scheduler queue corresponding to the specfied
// queue URL.
func (p *SqsQueueProvider) SendMessage(serviceID string, queueTopic string,
	msg *string) error {
	var (
		err      error
		queueUrl string
	)

	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	svcConfig := db.GetServiceConfig(serviceID)
	if svcConfig == nil {
		schedLogger.Error("Failed to retrieve service configuration!",
			zap.String("Service ID", serviceID),
		)
		return db.ErrInternalError
	}

	queueUrl, err = p.getQueueUrl(queueTopic, svcConfig.OwnerAwsAccount)
	if err != nil {
		schedLogger.Error("Failed to get the queue URL",
			zap.String("Queue topic", queueTopic),
			zap.Error(err),
		)
		return err
	}

	msg_output, err := p.gSQS.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: msg,
	})
	if err == nil {
		schedLogger.Info("Successfully sent message to service's queue!",
			zap.String("Service ID", serviceID),
			zap.String("Queue topic", queueTopic),
			zap.String("Queue URL", queueUrl),
			zap.String("Message ID", *msg_output.MessageId),
		)
	}
	return err
}

func (p *SqsQueueProvider) getQueueUrl(topicName string, ownerAwsAccount string) (string, error) {
	var (
		err       error
		urlResult *sqs.GetQueueUrlOutput
	)

	// Check if the queue URL has already been retrieved for this queue
	// topic name.
	queueUrl, ok := queueUrlLookupTable[topicName]
	if ok {
		return queueUrl, nil
	}

	// Figure out the queue URL and cache it in the lookup table.
	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	input := sqs.GetQueueUrlInput{
		QueueName: &topicName,
	}
	if ownerAwsAccount != "" {
		input.QueueOwnerAWSAccountId = &ownerAwsAccount
	}

	urlResult, err = p.gSQS.GetQueueUrl(ctx, &input)
	if err != nil {
		schedLogger.Error("Failed to get the queue URL for the requested queue topic!",
			zap.String("Queue topic", topicName),
			zap.String("Owner AWS account", ownerAwsAccount),
			zap.Error(err),
		)
		return "", err
	}

	queueUrlLookupTable[topicName] = *urlResult.QueueUrl
	return *urlResult.QueueUrl, nil
}

// Delete the specified message identified by its receipt handle from the
// specified scheduler queue.
func (p *SqsQueueProvider) deleteMessage(queueUrl string, receiptHandle string) error {
	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := p.gSQS.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: &receiptHandle,
	})
	return err
}
