package db

import (
	"testing"

	"github.com/google/uuid"
)

func toStringPtr(str string) *string {
	return &str
}

func toBytePtr(str string) *[]byte {
	bstr := []byte(str)
	return &bstr
}

func TestCreateTask(t *testing.T) {
	newTask, err := NewTask(toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do something"))
	if err != nil {
		t.Errorf("Failed to initialize task with error %v\n", err)
		return
	}

	err = newTask.CreateTask()
	if err != nil {
		t.Errorf("CreateTask failed with error %v\n", err)
		return
	}

	t.Logf("Created task with details : %+v\n", newTask)
}

func TestCreateTask_NoTenantID(t *testing.T) {
	newTask, err := NewTask(toStringPtr(""),
		toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do something"))
	if err != nil {
		t.Errorf("Failed to initialize task with error %v\n", err)
		return
	}

	err = newTask.CreateTask()
	assertEqual(t, err, ErrInvalidRequest)
}

func TestCreateTask_NoDeviceID(t *testing.T) {
	newTask, err := NewTask(toStringPtr(uuid.NewString()),
		toStringPtr(""),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do nothing good"))
	if err != nil {
		t.Logf("Failed to initialize task as expected! error %v\n", err)
		return
	}

	t.Errorf("Initialized task with no device ID! %v\n", newTask)
}
