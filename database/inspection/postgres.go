package inspection

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(db *sqlx.DB) *Postgres {
	return &Postgres{
		db: db,
	}
}

//go:embed sql/add.sql
var addSQL string

func (p *Postgres) Add(ctx context.Context, ins Inspection) (newIns Inspection, err error) {
	rows, err := p.db.NamedQueryContext(ctx, addSQL, ins)
	if err != nil {
		return Inspection{}, fmt.Errorf("p.db.NamedQueryContext: %w", err)
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	if !rows.Next() {
		return Inspection{}, errors.New("rows.Next == false")
	}

	err = rows.StructScan(&newIns)
	if err != nil {
		return Inspection{}, fmt.Errorf("rows.Scan: %w", err)
	}

	if err = rows.Err(); err != nil {
		return Inspection{}, fmt.Errorf("rows.Err: %w", err)
	}

	return newIns, err
}

//go:embed sql/add_attachment.sql
var addAttachmentSQL string

func (p *Postgres) AddAttachment(ctx context.Context, a Attachment) (newA Attachment, err error) {
	rows, err := p.db.NamedQueryContext(ctx, addAttachmentSQL, a)
	if err != nil {
		return Attachment{}, fmt.Errorf("p.db.NamedQueryContext: %w", err)
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	if !rows.Next() {
		return Attachment{}, errors.New("rows.Next == false")
	}

	err = rows.StructScan(&newA)
	if err != nil {
		return Attachment{}, fmt.Errorf("rows.Scan: %w", err)
	}

	if err = rows.Err(); err != nil {
		return Attachment{}, fmt.Errorf("rows.Err: %w", err)
	}

	return newA, err
}

//go:embed sql/get_by_task_id.sql
var getByTaskIDSQL string

func (p *Postgres) GetByTaskID(ctx context.Context, taskID int) (ins Inspection, err error) {
	err = p.db.GetContext(ctx, &ins, getByTaskIDSQL, taskID)
	if err != nil {
		return Inspection{}, fmt.Errorf("p.db.GetContext: %w", err)
	}

	return ins, nil
}

//go:embed sql/get_by_id.sql
var getByIDSQL string

func (p *Postgres) GetByID(ctx context.Context, id int) (ins Inspection, err error) {
	err = p.db.GetContext(ctx, &ins, getByIDSQL, id)
	if err != nil {
		return Inspection{}, fmt.Errorf("p.db.GetContext: %w", err)
	}

	return ins, nil
}

//go:embed sql/start_inspection.sql
var startInspectionSQL string

func (p *Postgres) StartInspection(ctx context.Context, taskID int) (ins Inspection, err error) {
	err = p.db.GetContext(ctx, &ins, startInspectionSQL, taskID)
	if err != nil {
		return Inspection{}, fmt.Errorf("p.db.GetContext: %w", err)
	}

	return ins, nil
}

//go:embed sql/finish_inspection.sql
var finishInspectionSQL string

func (p *Postgres) FinishInspection(ctx context.Context, ins Inspection) (newIns Inspection, err error) {
	rows, err := p.db.NamedQueryContext(ctx, finishInspectionSQL, ins)
	if err != nil {
		return Inspection{}, fmt.Errorf("p.db.NamedQueryContext: %w", err)
	}
	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	if !rows.Next() {
		return Inspection{}, errors.New("rows.Next == false")
	}

	err = rows.StructScan(&newIns)
	if err != nil {
		return Inspection{}, fmt.Errorf("rows.Scan: %w", err)
	}

	if err = rows.Err(); err != nil {
		return Inspection{}, fmt.Errorf("rows.Err: %w", err)
	}

	return newIns, err
}

//go:embed sql/get_previous_device_inspections.sql
var getPreviousDeviceInspectionsSQL string

func (p *Postgres) GetPreviousDeviceInspections(ctx context.Context, deviceID, inspectionID int) (devices []InspectedDevice, err error) {
	err = p.db.SelectContext(ctx, &devices, getPreviousDeviceInspectionsSQL, deviceID, inspectionID)
	if err != nil {
		return nil, fmt.Errorf("p.db.SelectContext: %w", err)
	}

	return devices, nil
}

//go:embed sql/add_device.sql
var addDeviceSQL string

func (p *Postgres) AddDevices(ctx context.Context, devices []InspectedDevice) error {
	_, err := p.db.NamedExecContext(ctx, addDeviceSQL, devices)
	if err != nil {
		return fmt.Errorf("p.db.NamedExecContext: %w", err)
	}

	return nil
}

//go:embed sql/add_seal.sql
var addSealSQL string

func (p *Postgres) AddSeals(ctx context.Context, seals []InspectedSeal) error {
	_, err := p.db.NamedExecContext(ctx, addSealSQL, seals)
	if err != nil {
		return fmt.Errorf("p.db.NamedExecContext: %w", err)
	}

	return nil
}
