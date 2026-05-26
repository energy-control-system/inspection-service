package inspection

import (
	"context"
	"database/sql"
	"io"
	"testing"

	clusterfile "inspection-service/cluster/file"
	clustertask "inspection-service/cluster/task"

	"github.com/sunshineOfficial/golib/goctx"
	"github.com/sunshineOfficial/golib/pagination"
)

type repositoryMock struct {
	inspections         []Inspection
	inspectionsByTaskID map[int]Inspection
	inspectionsByID     map[int]Inspection
}

func (m repositoryMock) GetAll(context.Context, pagination.Pagination) ([]Inspection, error) {
	return m.inspections, nil
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

type fileServiceMock struct {
	filesByID map[int]clusterfile.File
	gotIDs    []int
}

func (m *fileServiceMock) Upload(goctx.Context, string, io.Reader) (clusterfile.File, error) {
	return clusterfile.File{}, nil
}

func (m *fileServiceMock) GetByIDs(_ goctx.Context, ids []int, page pagination.Pagination) ([]clusterfile.File, error) {
	m.gotIDs = append([]int(nil), ids...)

	files := make([]clusterfile.File, 0, len(ids))
	for _, id := range ids {
		if f, ok := m.filesByID[id]; ok {
			files = append(files, f)
		}
	}

	return files, nil
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
		fileService: &fileServiceMock{},
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

func TestGetAllReturnsAttachmentFileURLs(t *testing.T) {
	fileService := &fileServiceMock{
		filesByID: map[int]clusterfile.File{
			70: {ID: 70, URL: "https://example.test/storage/photo.jpg"},
			80: {ID: 80, URL: "https://example.test/storage/act.docx"},
		},
	}

	service := &Service{
		repository: repositoryMock{
			inspections: []Inspection{
				{
					ID:     42,
					TaskID: 7,
					Attachments: []Attachment{
						{ID: 100, InspectionID: 42, FileID: 70, Type: AttachmentTypeDevicePhoto},
					},
				},
				{
					ID:     43,
					TaskID: 8,
					Attachments: []Attachment{
						{ID: 101, InspectionID: 43, FileID: 80, Type: AttachmentTypeAct},
					},
				},
			},
		},
		fileService: fileService,
	}

	got, err := service.GetAll(goctx.Wrap(context.Background()), pagination.Pagination{})
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}

	if got[0].Attachments[0].FileURL != "https://example.test/storage/photo.jpg" {
		t.Fatalf("got[0].Attachments[0].FileURL = %q, want %q", got[0].Attachments[0].FileURL, "https://example.test/storage/photo.jpg")
	}
	if got[1].Attachments[0].FileURL != "https://example.test/storage/act.docx" {
		t.Fatalf("got[1].Attachments[0].FileURL = %q, want %q", got[1].Attachments[0].FileURL, "https://example.test/storage/act.docx")
	}
	if len(fileService.gotIDs) != 2 || fileService.gotIDs[0] != 70 || fileService.gotIDs[1] != 80 {
		t.Fatalf("fileService.gotIDs = %+v, want [70 80]", fileService.gotIDs)
	}
}

func TestGetByIDReturnsInspection(t *testing.T) {
	fileService := &fileServiceMock{
		filesByID: map[int]clusterfile.File{
			70: {ID: 70, URL: "https://example.test/storage/photo.jpg"},
		},
	}

	service := &Service{
		repository: repositoryMock{
			inspectionsByID: map[int]Inspection{
				42: {
					ID:     42,
					TaskID: 7,
					Status: StatusInWork,
					Attachments: []Attachment{
						{ID: 100, InspectionID: 42, FileID: 70, Type: AttachmentTypeDevicePhoto},
					},
				},
			},
		},
		fileService: fileService,
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
	if got.Attachments[0].FileURL != "https://example.test/storage/photo.jpg" {
		t.Fatalf("got.Attachments[0].FileURL = %q, want %q", got.Attachments[0].FileURL, "https://example.test/storage/photo.jpg")
	}
	if len(fileService.gotIDs) != 1 || fileService.gotIDs[0] != 70 {
		t.Fatalf("fileService.gotIDs = %+v, want [70]", fileService.gotIDs)
	}
}
