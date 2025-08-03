package sqs_provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/mqtt"
	"go.uber.org/zap"
)

// Watch the scheduler dispatch queue for new requests at the configured watch
// interval.
func (p *SqsQueueProvider) WatchDispatchQueue() {
	schedLogger.Info("Dispatch Queue Watcher: Watching the scheduler dispatch queue for requests!",
		zap.String("Queue name:", p.queueConfig.DispatchQueueName),
		zap.Int32("Watch delay:", p.queueConfig.WatchDelay),
	)

	for {
		// Check if the queue manager needs to shut down. If so, stop processing
		// messages.
		if p.gCtx.Err() != nil {
			schedLogger.Info("Received a message to shutdown. Processing of scheduler dispatch events is stopped!")
			break
		}

		// Look for requests on the scheduler dispatch queue.
		p.processSchedulerDispatchRequest()
	}
}

// Retrieve a single message from the scheduler dispatch queue and dispatch it
// for processing. This function will process messages from the dispatch queue
// one at a time until there are no more messages on the queue.
func (p *SqsQueueProvider) processSchedulerDispatchRequest() {

	// Receive a single message from the scheduler dispatch queue.
	taskInfo, payload, receiptHandle, err := p.receiveDispatchQueueMessage()
	if err != nil {
		schedLogger.Error("Failed to receive message from scheduler dispatch queue!",
			zap.Error(err))
		return
	}

	// If there are no messages, we have nothing to do.
	if taskInfo == nil {
		return
	}

	// Dispatch the received message to the MQTT broker for delivery to the
	// target device.
	err = mqtt.SendTaskToBroker(
		mqtt.GetMqttTopicForDeviceTask(taskInfo.DeviceId, taskInfo.ServiceId),
		payload)
	if err != nil {
		schedLogger.Error("Failed to process message on scheduler dispatch queue",
			zap.String("Task ID: ", taskInfo.TaskId),
			zap.String("Device ID: ", taskInfo.DeviceId),
			zap.Error(err),
		)
		return
	}

	err = db.MarkTaskDispatched(taskInfo)
	if err != nil {
		schedLogger.Error("Failed to update task status to dispatched!",
			zap.String("Task ID: ", taskInfo.TaskId),
			zap.String("Device ID: ", taskInfo.DeviceId),
			zap.Error(err),
		)
		// This may result in the same message being delivered multiple times.
		return
	}

	// Delete the processed message from the scheduler dispatch queue.
	err = p.deleteMessage(p.schedulerDispatchQueueUrl, receiptHandle)
	if err != nil {
		schedLogger.Error("Failed to remove message from scheduler dispatch queue!",
			zap.String("Task ID: ", taskInfo.TaskId),
			zap.String("Device ID: ", taskInfo.DeviceId),
			zap.Error(err),
		)
	}
}

// Receive a single message on the scheduler dispatch queue. Unmarshal the request
// into a db.Task structure. Return the task request and message handle which can
// be used to acknowledge processing of the request.
func (p *SqsQueueProvider) receiveDispatchQueueMessage() (*pb.ServiceMessage, *[]byte,
	string, error) {
	msg, err := p.receiveMessage(p.schedulerDispatchQueueUrl)
	if err != nil {
		schedLogger.Error("Error receiving message from scheduler input queue",
			zap.String("Queue URL:", p.schedulerDispatchQueueUrl),
			zap.Error(err))
		return nil, nil, "", err
	}

	if msg == nil {
		return nil, nil, "", nil
	}

	// Unmarshal the request received at the scheduler dispatch queue.
	request, encodedRequest, err := db.UnmarshallServiceMessage(msg.Messages[0].Body)
	if err != nil {
		schedLogger.Error("Failed to unmarshal request received from scheduler dispatch queue!",
			zap.String("Queue URL:", p.schedulerDispatchQueueUrl),
			zap.Error(err),
		)

		// Failed to unmarshal the request - delete it from the dispatch queue.
		_ = p.deleteMessage(p.schedulerDispatchQueueUrl, *msg.Messages[0].ReceiptHandle)
		return nil, nil, "", err
	}

	return request, encodedRequest, *msg.Messages[0].ReceiptHandle, nil
}

func (p *SqsQueueProvider) SendDispatchQueueMessage(msg *string) error {
	schedLogger.Info("Sending message to the scheduler dispatch queue!",
		zap.String("Message:", *msg),
	)

	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	_, err := p.gSQS.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &p.schedulerDispatchQueueUrl,
		MessageBody: msg,
	})
	return err
}
