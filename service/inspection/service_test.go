package inspection

import (
	"context"
	"database/sql"
	"testing"

	clustertask "inspection-service/cluster/task"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/pagination"
)

type repositoryMock struct {
	inspectionsByTaskID map[int]Inspection
	inspectionsByID     map[int]Inspection
}

func (m repositoryMock) GetAll(context.Context, pagination.Pagination) ([]Inspection, error) {
	return nil, nil
}

func (m repositoryMock) GetByTaskID(_ context.Context, taskID int) (Inspection, error) {
	ins, ok := m.inspectionsByTaskID[taskID]
	if !ok {
		return Inspection{}, sql.ErrNoRows
	}

	return ins, nil
}

func (m repositoryMock) AddAttachment(context.Context, int, int, AttachmentType) (Attachment, error) {
	return Attachment{}, nil
}

func (m repositoryMock) GetByID(_ context.Context, id int) (Inspection, error) {
	ins, ok := m.inspectionsByID[id]
	if !ok {
		return Inspection{}, sql.ErrNoRows
	}

	return ins, nil
}

func (m repositoryMock) GetPreviousDeviceInspections(context.Context, int, int) ([]InspectedDevice, error) {
	return nil, nil
}

func (m repositoryMock) AddInspectedDevices(context.Context, int, []InspectedDeviceRequest) error {
	return nil
}

func (m repositoryMock) StartInspection(context.Context, int) (Inspection, error) {
	return Inspection{}, nil
}

func (m repositoryMock) FinishInspection(context.Context, FinishInspectionRequest) (Inspection, error) {
	return Inspection{}, nil
}

type taskServiceMock struct {
	tasksByBrigadeID map[int][]clustertask.Task
	gotPage          pagination.Pagination
}

func (m taskServiceMock) GetTaskByID(goctx.Context, int) (clustertask.Task, error) {
	return clustertask.Task{}, nil
}

func (m *taskServiceMock) GetTasksByBrigade(_ goctx.Context, brigadeID int, page pagination.Pagination) ([]clustertask.Task, error) {
	m.gotPage = page

	return m.tasksByBrigadeID[brigadeID], nil
}

func TestGetByBrigadeReturnsInspectionsForBrigadeTasks(t *testing.T) {
	taskService := &taskServiceMock{
		tasksByBrigadeID: map[int][]clustertask.Task{
			7: {
				{ID: 10},
				{ID: 20},
				{ID: 30},
			},
		},
	}

	service := &Service{
		repository: repositoryMock{
			inspectionsByTaskID: map[int]Inspection{
				10: {ID: 100, TaskID: 10},
				30: {ID: 300, TaskID: 30},
			},
		},
		taskService: taskService,
	}

	page := pagination.Pagination{Limit: 2, Offset: 4}
	got, err := service.GetByBrigade(goctx.Wrap(context.Background()), 7, page)
	if err != nil {
		t.Fatalf("GetByBrigade returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].ID != 100 {
		t.Fatalf("got[0].ID = %d, want 100", got[0].ID)
	}
	if got[1].ID != 300 {
		t.Fatalf("got[1].ID = %d, want 300", got[1].ID)
	}
	if taskService.gotPage != page {
		t.Fatalf("taskService.gotPage = %+v, want %+v", taskService.gotPage, page)
	}
}

func TestGetByIDReturnsInspection(t *testing.T) {
	service := &Service{
		repository: repositoryMock{
			inspectionsByID: map[int]Inspection{
				42: {ID: 42, TaskID: 7, Status: StatusInWork},
			},
		},
	}

	got, err := service.GetByID(goctx.Wrap(context.Background()), 42)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}

	if got.ID != 42 {
		t.Fatalf("got.ID = %d, want 42", got.ID)
	}
	if got.TaskID != 7 {
		t.Fatalf("got.TaskID = %d, want 7", got.TaskID)
	}
	if got.Status != StatusInWork {
		t.Fatalf("got.Status = %d, want %d", got.Status, StatusInWork)
	}
}
