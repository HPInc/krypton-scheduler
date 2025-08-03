package sqs_provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

func (p *SqsQueueProvider) SendDcmInputQueueMessage(msg *string) error {
	schedLogger.Info("Sending message to the DCM input queue!",
		zap.String("Message:", *msg),
	)

	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := p.gSQS.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &p.dcmInputQueueUrl,
		MessageBody: msg,
	})
	return err
}
