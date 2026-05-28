package inspection

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"inspection-service/service/inspection"

	"github.com/jmoiron/sqlx"
	"github.com/sunshineOfficial/golib/pagination"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

//go:embed sql/get_all.sql
var getAllSQL string

func (r *Repository) GetAll(ctx context.Context, page pagination.Pagination, sort inspection.SortDirection) ([]inspection.Inspection, error) {
	var inspections []Inspection
	err := r.db.SelectContext(ctx, &inspections, getAllSQL, sort, page.LimitArg(), page.Offset)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	err = r.attachAttachments(ctx, inspections)
	if err != nil {
		return nil, fmt.Errorf("attach attachments: %w", err)
	}

	return MapSliceFromDB(inspections), nil
}

//go:embed sql/get_by_task_id.sql
var getByTaskIDSQL string

func (r *Repository) GetByTaskID(ctx context.Context, taskID int) (inspection.Inspection, error) {
	var ins Inspection
	err := r.db.GetContext(ctx, &ins, getByTaskIDSQL, taskID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("r.db.GetContext: %w", err)
	}

	result := MapFromDB(ins)
	result.Attachments, err = r.getAttachmentsByInspectionID(ctx, ins.ID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("get attachments by inspection id: %w", err)
	}

	result.InspectedDevices, err = r.getDevicesByInspectionID(ctx, ins.ID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("get devices by inspection id: %w", err)
	}

	return result, nil
}

//go:embed sql/get_devices_by_inspection_id.sql
var getDevicesByInspectionIDSQL string

func (r *Repository) getDevicesByInspectionID(ctx context.Context, inspectionID int) ([]inspection.InspectedDevice, error) {
	var devices []InspectedDevice
	err := r.db.SelectContext(ctx, &devices, getDevicesByInspectionIDSQL, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	return MapInspectedDevicesSliceFromDB(devices), nil
}

//go:embed sql/get_attachments_by_inspection_ids.sql
var getAttachmentsByInspectionIDsSQL string

func (r *Repository) getAttachmentsByInspectionID(ctx context.Context, inspectionID int) ([]inspection.Attachment, error) {
	attachments, err := r.getAttachmentsByInspectionIDs(ctx, []int{inspectionID})
	if err != nil {
		return nil, err
	}

	return MapAttachmentsSliceFromDB(attachments), nil
}

func (r *Repository) attachAttachments(ctx context.Context, inspections []Inspection) error {
	if len(inspections) == 0 {
		return nil
	}

	inspectionIDs := make([]int, 0, len(inspections))
	for _, ins := range inspections {
		inspectionIDs = append(inspectionIDs, ins.ID)
	}

	attachments, err := r.getAttachmentsByInspectionIDs(ctx, inspectionIDs)
	if err != nil {
		return err
	}

	attachmentsByInspectionID := make(map[int][]Attachment, len(inspections))
	for _, attachment := range attachments {
		attachmentsByInspectionID[attachment.InspectionID] = append(attachmentsByInspectionID[attachment.InspectionID], attachment)
	}

	for i := range inspections {
		inspections[i].Attachments = attachmentsByInspectionID[inspections[i].ID]
	}

	return nil
}

func (r *Repository) getAttachmentsByInspectionIDs(ctx context.Context, inspectionIDs []int) ([]Attachment, error) {
	query, args, err := sqlx.In(getAttachmentsByInspectionIDsSQL, inspectionIDs)
	if err != nil {
		return nil, fmt.Errorf("sqlx.In: %w", err)
	}

	query = r.db.Rebind(query)

	var attachments []Attachment
	err = r.db.SelectContext(ctx, &attachments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	return attachments, nil
}

//go:embed sql/add_attachment.sql
var addAttachmentSQL string

func (r *Repository) AddAttachment(ctx context.Context, inspectionID, fileID int, attachmentType inspection.AttachmentType) (inspection.Attachment, error) {
	var a Attachment
	err := r.db.GetContext(ctx, &a, addAttachmentSQL, inspectionID, attachmentType, fileID)
	if err != nil {
		return inspection.Attachment{}, fmt.Errorf("r.db.GetContext: %w", err)
	}

	return MapAttachmentFromDB(a), nil
}

//go:embed sql/get_by_id.sql
var getByIDSQL string

func (r *Repository) GetByID(ctx context.Context, id int) (inspection.Inspection, error) {
	var ins Inspection
	err := r.db.GetContext(ctx, &ins, getByIDSQL, id)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("r.db.GetContext: %w", err)
	}

	result := MapFromDB(ins)
	result.Attachments, err = r.getAttachmentsByInspectionID(ctx, ins.ID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("get attachments by inspection id: %w", err)
	}

	return result, nil
}

//go:embed sql/get_previous_device_inspections.sql
var getPreviousDeviceInspectionsSQL string

func (r *Repository) GetPreviousDeviceInspections(ctx context.Context, inspectionID, deviceID int) ([]inspection.InspectedDevice, error) {
	var devices []InspectedDevice
	err := r.db.SelectContext(ctx, &devices, getPreviousDeviceInspectionsSQL, deviceID, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("r.db.SelectContext: %w", err)
	}

	return MapInspectedDevicesSliceFromDB(devices), nil
}

//go:embed sql/add_device.sql
var addDeviceSQL string

//go:embed sql/add_seal.sql
var addSealSQL string

func (r *Repository) AddInspectedDevices(ctx context.Context, inspectionID int, requests []inspection.InspectedDeviceRequest) error {
	devices, seals := MapInspectedDeviceRequestsSliceToDB(requests, inspectionID)

	_, err := r.db.NamedExecContext(ctx, addDeviceSQL, devices)
	if err != nil {
		return fmt.Errorf("add devices: %w", err)
	}

	_, err = r.db.NamedExecContext(ctx, addSealSQL, seals)
	if err != nil {
		return fmt.Errorf("add seals: %w", err)
	}

	return nil
}

//go:embed sql/start_inspection.sql
var startInspectionSQL string

func (r *Repository) StartInspection(ctx context.Context, taskID int) (inspection.Inspection, error) {
	var ins Inspection
	err := r.db.GetContext(ctx, &ins, startInspectionSQL, taskID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("r.db.GetContext: %w", err)
	}

	result := MapFromDB(ins)
	result.Attachments = []inspection.Attachment{}

	return result, nil
}

//go:embed sql/finish_inspection.sql
var finishInspectionSQL string

func (r *Repository) FinishInspection(ctx context.Context, request inspection.FinishInspectionRequest) (inspection.Inspection, error) {
	dbRequest := MapFinishInspectionRequestToDB(request)

	rows, err := r.db.NamedQueryContext(ctx, finishInspectionSQL, dbRequest)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("r.db.NamedQueryContext: %w", err)
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	if !rows.Next() {
		return inspection.Inspection{}, errors.New("rows.Next == false")
	}

	var ins Inspection
	err = rows.StructScan(&ins)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("rows.Scan: %w", err)
	}

	if err = rows.Err(); err != nil {
		return inspection.Inspection{}, fmt.Errorf("rows.Err: %w", err)
	}

	result := MapFromDB(ins)
	result.Attachments, err = r.getAttachmentsByInspectionID(ctx, ins.ID)
	if err != nil {
		return inspection.Inspection{}, fmt.Errorf("get attachments by inspection id: %w", err)
	}

	return result, err
}
