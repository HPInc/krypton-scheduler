package sqs_provider

import (
	"context"
	"errors"
	"net/url"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	schedLogger *zap.Logger

	queueUrlLookupTable map[string]string
)

const (
	queueUrlFormat          = "%s/queue/%s"
	awsOperationTimeout     = time.Second * 5
	awsSqsVisibilityTimeout = 60
)

type SqsQueueProvider struct {
	// Connection to the SQS queue.
	gSQS *sqs.Client

	gCtx context.Context

	// queue URLs
	schedulerInputQueueUrl    string
	schedulerDispatchQueueUrl string
	dcmInputQueueUrl          string

	// Queue configuration.
	queueConfig *config.QueueMgrConfig

	// Handler function to process requests received at the scheduler's
	// input queue.
	inputEventHandlerFunc common.InputEventHandlerFunc
}

func NewSqsProvider() *SqsQueueProvider {
	return &SqsQueueProvider{}
}

// Initialize the SQS queue provider and create an SQS client to be used for
// connections to SQS.
func (p *SqsQueueProvider) Init(logger *zap.Logger,
	cfgMgr *config.ConfigMgr) error {
	schedLogger = logger
	p.queueConfig = cfgMgr.GetQueueMgrConfig()

	p.gCtx = context.Background()

	// Initialize a new SQS client.
	err := p.initSQSApiClient(p.queueConfig)
	if err != nil {
		schedLogger.Error("Failed to initialize an SQS client using the configuration!",
			zap.Error(err),
		)
		return err
	}

	queueUrlLookupTable = make(map[string]string)
	return nil
}

type resolverV2 struct {
	// Custom SQS endpoint, if configured.
	endpoint string
}

// make endpoint connection for transparent runs in local as well as cloud.
// Specify endpoint explicitly for local runs; cloud runs will load default
// config automatically. settings.Endpoint will not be set for cloud runs
func (r *resolverV2) ResolveEndpoint(ctx context.Context, params sqs.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.endpoint != "" {
		uri, err := url.Parse(r.endpoint)
		return smithyendpoints.Endpoint{
			URI: *uri,
		}, err
	}

	// delegate back to the default v2 resolver otherwise
	return sqs.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// Create a new sqs client using the passed in configuration.
func (p *SqsQueueProvider) initSQSApiClient(settings *config.QueueMgrConfig) error {
	ctx, cancelFunc := context.WithTimeout(p.gCtx, awsOperationTimeout)
	defer cancelFunc()

	// Initialize an AWS session to access the queue.
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		schedLogger.Error("Failed to load default configuration for the SQS provider.",
			zap.Error(err),
		)
		return err
	}

	// Initialize the SQS client using this session.
	p.gSQS = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.EndpointResolverV2 = &resolverV2{endpoint: settings.Endpoint}
	})
	if p.gSQS == nil {
		schedLogger.Error("Failed to create a new queue client!",
			zap.Error(err),
		)
		return errors.New("could not create SQS client")
	}

	// Configure the queue URLs for the scheduler input & dispatch queues.
	urlResult, err := p.gSQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &p.queueConfig.InputQueueName,
	})
	if err != nil {
		schedLogger.Error("Failed to get the scheduler input queue URL!",
			zap.Error(err),
		)
		return err
	}
	p.schedulerInputQueueUrl = *urlResult.QueueUrl

	urlResult, err = p.gSQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &p.queueConfig.DispatchQueueName,
	})
	if err != nil {
		schedLogger.Error("Failed to get the scheduler dispatch queue URL!",
			zap.Error(err),
		)
		return err
	}
	p.schedulerDispatchQueueUrl = *urlResult.QueueUrl

	// Determine the queue URL for the DCM input queue.
	urlResult, err = p.gSQS.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &p.queueConfig.DcmInputQueueName,
	})
	if err != nil {
		schedLogger.Error("Failed to get the DCM input queue URL!",
			zap.Error(err),
		)
		return err
	}
	p.dcmInputQueueUrl = *urlResult.QueueUrl

	return nil
}

// Shutdown the SQS queue provider.
func (p *SqsQueueProvider) Shutdown() {
	// Cancel the main context so the goroutine processing the scheduler input
	// and dispatch queues can stop.
	if p.gCtx != nil {
		p.gCtx.Done()
	}
}
