package inspection

import (
	"context"
	"inspection-service/cluster/analyzer"
	"inspection-service/cluster/brigade"
	"inspection-service/cluster/file"
	"inspection-service/cluster/subscriber"
	"inspection-service/cluster/task"
	"io"

	"github.com/sunshineOfficial/golib/goctx"
)

type Repository interface {
	GetByTaskID(ctx context.Context, taskID int) (Inspection, error)
	AddAttachment(ctx context.Context, inspectionID, fileID int, attachmentType AttachmentType) (Attachment, error)
	GetByID(ctx context.Context, id int) (Inspection, error)
	GetPreviousDeviceInspections(ctx context.Context, inspectionID, deviceID int) ([]InspectedDevice, error)
	AddInspectedDevices(ctx context.Context, inspectionID int, requests []InspectedDeviceRequest) error
	StartInspection(ctx context.Context, taskID int) (Inspection, error)
	FinishInspection(ctx context.Context, request FinishInspectionRequest) (Inspection, error)
}

type AnalyzerService interface {
	ProcessImage(ctx goctx.Context, fileName string, image io.Reader) (analyzer.ProcessImageResponse, error)
}

type SubscriberService interface {
	GetObjectExtendedByID(ctx goctx.Context, id int) (subscriber.ObjectExtended, error)
	GetObjectExtendedByDevice(ctx goctx.Context, deviceID int) (subscriber.ObjectExtended, error)
	GetObjectExtendedBySeal(ctx goctx.Context, sealID int) (subscriber.ObjectExtended, error)
}

type FileService interface {
	Upload(ctx goctx.Context, fileName string, file io.Reader) (file.File, error)
}

type TaskService interface {
	GetTaskByID(ctx goctx.Context, id int) (task.Task, error)
}

type BrigadeService interface {
	GetBrigadeByID(ctx goctx.Context, id int) (brigade.Brigade, error)
}
