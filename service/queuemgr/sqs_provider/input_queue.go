package sqs_provider

import (
	b64 "encoding/base64"

	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Watch the scheduler input queue for new requests at the configured watch
// interval.
func (p *SqsQueueProvider) WatchInputQueue(onInputEvent common.InputEventHandlerFunc) {
	p.inputEventHandlerFunc = onInputEvent

	schedLogger.Info("Input Queue Watcher: Watching the scheduler input queue for requests!",
		zap.String("Queue name:", p.queueConfig.InputQueueName),
		zap.Int32("Watch delay:", p.queueConfig.WatchDelay),
	)
	for {
		// Check if the queue manager needs to shut down. If so, stop processing
		// messages.
		if p.gCtx.Err() != nil {
			schedLogger.Info("Input Queue Watcher: No longer watching scheduler input queue.")
			break
		}

		// Look for requests on the scheduler input queue.
		p.processSchedulerInputQueueMessage()
	}
}

// Retrieve a single message from the scheduler input queue and dispatch it
// for processing. This function will process messages from the input queue
// one at a time until there are no more messages on the queue.
func (p *SqsQueueProvider) processSchedulerInputQueueMessage() {
	// Receive a single schedule task request from the scheduler input queue.
	taskRequest, receiptHandle, err := p.receiveInputQueueMessage()
	if err != nil {
		schedLogger.Error("Failed to receive message from scheduler input queue!",
			zap.Error(err))
		return
	}

	// If there are no messages, we have nothing to do.
	if taskRequest == nil {
		return
	}

	// Dispatch the received message for processing.
	_, err = p.inputEventHandlerFunc(taskRequest, common.SchedulerRequestSourceEvent)
	if err != nil {
		schedLogger.Error("Failed to process message on scheduler input queue",
			zap.Error(err),
		)
		// Fall through to delete the message from the input queue.
	}

	// Delete the processed message from the scheduler input queue.
	p.disposeBadInputQueueRequest(receiptHandle)
}

// Receive a single message on the scheduler input queue. Unmarshal the request
// into a ScheduledTaskRequest structure using protobuf. Return the task request
// and message handle which can be used to acknowledge processing of the request.
func (p *SqsQueueProvider) receiveInputQueueMessage() (*pb.CreateScheduledTaskRequest,
	string, error) {
	msg, err := p.receiveMessage(p.schedulerInputQueueUrl)
	if err != nil {
		schedLogger.Error("Error receiving message from scheduler input queue",
			zap.String("Queue URL", p.schedulerInputQueueUrl),
			zap.Error(err))
		return nil, "", err
	}

	if msg == nil {
		return nil, "", nil
	}

	// Base 64 decode the packet from string format into a protobuf encoded
	// byte stream
	packetBytes, err := b64.StdEncoding.DecodeString(*msg.Messages[0].Body)
	if err != nil {
		schedLogger.Error("Failed to base64 decode the message at the scheduler input queue",
			zap.Error(err),
		)
		// Failed to decode the request - delete it from the input queue.
		p.disposeBadInputQueueRequest(*msg.Messages[0].ReceiptHandle)
		return nil, "", err
	}

	// Unmarshal the request received at the scheduler input queue.
	var request pb.CreateScheduledTaskRequest
	err = proto.Unmarshal(packetBytes, &request)
	if err != nil {
		schedLogger.Error("Failed to unmarshal request received from scheduler input queue!",
			zap.String("Queue URL", p.schedulerInputQueueUrl),
			zap.Error(err),
		)

		// Failed to unmarshal the request - delete it from the input queue.
		p.disposeBadInputQueueRequest(*msg.Messages[0].ReceiptHandle)
		return nil, "", err
	}

	return &request, *msg.Messages[0].ReceiptHandle, nil
}

func (p *SqsQueueProvider) disposeBadInputQueueRequest(receiptHandle string) {
	// Failed to unmarshal the request - delete it from the input queue.
	err := p.deleteMessage(p.schedulerInputQueueUrl, receiptHandle)
	if err != nil {
		schedLogger.Error("Failed to delete message from scheduler input queue!",
			zap.String("Queue URL", p.schedulerInputQueueUrl),
			zap.Error(err),
		)
	}
}
