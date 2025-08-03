package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"github.com/hpinc/krypton-scheduler/service/scheduler"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func init() {
	// Initialize logging for the test run.
	logger, err := zap.NewProduction(zap.AddCaller())
	if err != nil {
		fmt.Println("Failed to intialize structured logging for the RPC test server!")
		os.Exit(2)
	}

	// Read and parse the configuration file.
	cfgMgr := config.NewConfigMgr(logger, "Scheduler test service")
	if !cfgMgr.Load(true) {
		fmt.Println("Failed to load the configuration! Exiting ...")
		os.Exit(2)
	}

	err = db.Init(logger, cfgMgr)
	if err != nil {
		fmt.Printf("Failed to initialize the database with error %v\n", err)
		os.Exit(2)
	}

	err = scheduler.Init(logger, cfgMgr)
	if err != nil {
		fmt.Printf("Failed to initialize the scheduler engine with error %v\n", err)
		os.Exit(2)
	}

	err = queuemgr.Init(logger, cfgMgr, scheduler.ScheduleRequestHandlerFunc)
	if err != nil {
		fmt.Printf("Failed to initialize the queue manager with error %v\n", err)
		os.Exit(2)
	}

	InitTestServer(logger, cfgMgr)
}

func TestCreateScheduledTask(t *testing.T) {
	wr := httptest.NewRecorder()

	request := &pb.CreateScheduledTaskRequest{
		Version:       1,
		ServiceId:     "hpcem",
		DeviceIds:     []string{uuid.NewString()},
		ConsignmentId: uuid.NewString(),
		TenantId:      uuid.NewString(),
		Schedule:      "Every 1h",
		MessageType:   "CCS.GetConfig",
		Payload:       []byte("Do something epic!"),
	}
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		t.Errorf("Failed to encode scheduled task request! Error: %v\n", err)
		return
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks",
		bytes.NewReader(requestBytes))
	req.Header.Add(headerContentType, contentTypeProtobuf)

	CreateTaskHandler(wr, req)
	if wr.Code != http.StatusCreated {
		t.Errorf("Failed to create task! Returned code %v\n", wr.Code)
		return
	}

	t.Logf("Create task successful! Response: %v\n", wr)
}

func TestCreateBroadcastScheduledTask(t *testing.T) {
	wr := httptest.NewRecorder()

	request := &pb.CreateScheduledTaskRequest{
		Version:       1,
		ServiceId:     "hpcem",
		DeviceIds:     []string{common.BroadcastDeviceID},
		ConsignmentId: uuid.NewString(),
		TenantId:      uuid.NewString(),
		MessageType:   "CCS.GetConfig",
		Payload:       []byte("All devices report status"),
	}
	requestBytes, err := proto.Marshal(request)
	if err != nil {
		t.Errorf("Failed to encode scheduled task request! Error: %v\n", err)
		return
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks",
		bytes.NewReader(requestBytes))
	req.Header.Add(headerContentType, contentTypeProtobuf)

	CreateTaskHandler(wr, req)
	if wr.Code != http.StatusCreated {
		t.Errorf("Failed to create task! Returned code %v\n", wr.Code)
		return
	}

	t.Logf("Create task successful! Response: %v\n", wr)
}
