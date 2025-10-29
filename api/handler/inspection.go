package handler

import (
	"fmt"
	"inspection-service/service/inspection"
	"net/http"
	"strconv"

	"github.com/sunshineOfficial/golib/gohttp/gorouter"
)

func GetAllInspections(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		response, err := s.GetAll(c.Ctx())
		if err != nil {
			return fmt.Errorf("failed to get all inspections: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}

type taskIDVars struct {
	TaskID int `path:"taskID"`
}

func GetInspectionByTaskID(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		var vars taskIDVars
		if err := c.Vars(&vars); err != nil {
			return fmt.Errorf("failed to read task id: %w", err)
		}

		response, err := s.GetByTaskID(c.Ctx(), vars.TaskID)
		if err != nil {
			return fmt.Errorf("failed to get inspection by task id: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}

type inspectionIDVars struct {
	ID int `path:"id"`
}

func AttachPhotoToInspection(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		var vars inspectionIDVars
		if err := c.Vars(&vars); err != nil {
			return fmt.Errorf("failed to read inspection id: %w", err)
		}

		files, err := c.FormFiles("Photo")
		if err != nil {
			return fmt.Errorf("parse photo from form: %w", err)
		}
		if len(files) != 1 {
			return fmt.Errorf("got %d photos, expected 1", len(files))
		}

		attachmentTypes, err := c.FormValues("AttachmentType")
		if err != nil {
			return fmt.Errorf("parse attachment type from form: %w", err)
		}
		if len(attachmentTypes) != 1 {
			return fmt.Errorf("got %d attachment types, expected 1", len(attachmentTypes))
		}

		attachmentTypeRaw, err := strconv.Atoi(attachmentTypes[0])
		if err != nil {
			return fmt.Errorf("invalid attachment type: %s", attachmentTypes[0])
		}

		attachmentType := inspection.AttachmentType(attachmentTypeRaw)

		var deviceID, sealID int
		switch attachmentType {
		case inspection.AttachmentTypeDevicePhoto:
			deviceIDs, fErr := c.FormValues("DeviceID")
			if fErr != nil {
				return fmt.Errorf("parse device id from form: %w", fErr)
			}
			if len(deviceIDs) != 1 {
				return fmt.Errorf("got %d device ids, expected 1", len(deviceIDs))
			}

			deviceID, fErr = strconv.Atoi(deviceIDs[0])
			if fErr != nil {
				return fmt.Errorf("invalid device id: %s", deviceIDs[0])
			}

		case inspection.AttachmentTypeSealPhoto:
			sealIDs, fErr := c.FormValues("SealID")
			if fErr != nil {
				return fmt.Errorf("parse seal id from form: %w", fErr)
			}
			if len(sealIDs) != 1 {
				return fmt.Errorf("got %d seal ids, expected 1", len(sealIDs))
			}

			sealID, fErr = strconv.Atoi(sealIDs[0])
			if fErr != nil {
				return fmt.Errorf("invalid seal id: %s", sealIDs[0])
			}

		default:
			return fmt.Errorf("invalid attachment type: %d", attachmentType)
		}

		response, err := s.AttachPhoto(c.Ctx(), c.Log().WithTags("AttachPhoto"), inspection.AttachPhotoRequest{
			InspectionID: vars.ID,
			Type:         attachmentType,
			DeviceID:     deviceID,
			SealID:       sealID,
			FileHeader:   files[0],
		})
		if err != nil {
			return fmt.Errorf("failed to attach photo to inspection: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}

func FinishInspection(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		var vars inspectionIDVars
		if err := c.Vars(&vars); err != nil {
			return fmt.Errorf("failed to read id: %w", err)
		}

		var request inspection.FinishInspectionRequest
		if err := c.ReadJson(&request); err != nil {
			return fmt.Errorf("failed to read finish inspection request: %w", err)
		}

		request.ID = vars.ID

		response, err := s.FinishInspection(c.Ctx(), c.Log().WithTags("FinishInspection"), request)
		if err != nil {
			return fmt.Errorf("failed to finish inspection: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}
