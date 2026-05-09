package handler

import (
	"fmt"
	"inspection-service/service/inspection"
	"net/http"
	"strconv"

	"github.com/sunshineOfficial/golib/gohttp/gorouter"
	"github.com/sunshineOfficial/golib/pagination"
)

// GetAllInspections godoc
// @Summary List inspections
// @Description Returns all inspections.
// @Tags inspections
// @Produce json
// @Param limit query int false "Maximum number of items to return; 0 means no limit"
// @Param offset query int false "Number of items to skip"
// @Success 200 {array} inspection.Inspection
// @Failure 400 {object} gorouter.ErrorResponse
// @Failure 500 {object} gorouter.ErrorResponse
// @Router /inspections [get]
func GetAllInspections(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		var vars pagination.Pagination
		if err := c.Vars(&vars); err != nil {
			return fmt.Errorf("failed to read pagination: %w", err)
		}

		response, err := s.GetAll(c.Ctx(), vars)
		if err != nil {
			return fmt.Errorf("failed to get all inspections: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}

type taskIDVars struct {
	TaskID int `path:"taskID"`
}

// GetInspectionByTaskID godoc
// @Summary Get inspection by task ID
// @Description Returns an inspection linked to a task.
// @Tags inspections
// @Produce json
// @Param taskID path int true "Task ID"
// @Success 200 {object} inspection.Inspection
// @Failure 400 {object} gorouter.ErrorResponse
// @Failure 404 {object} gorouter.ErrorResponse
// @Failure 500 {object} gorouter.ErrorResponse
// @Router /inspections/task/{taskID} [get]
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

type brigadeIDVars struct {
	BrigadeID int `path:"brigadeID"`
}

// GetInspectionsByBrigade godoc
// @Summary List brigade inspections
// @Description Returns inspections linked to tasks assigned to a brigade.
// @Tags inspections
// @Produce json
// @Param brigadeID path int true "Brigade ID"
// @Param limit query int false "Maximum number of items to return; 0 means no limit"
// @Param offset query int false "Number of items to skip"
// @Success 200 {array} inspection.Inspection
// @Failure 400 {object} gorouter.ErrorResponse
// @Failure 500 {object} gorouter.ErrorResponse
// @Router /inspections/brigades/{brigadeID} [get]
func GetInspectionsByBrigade(s *inspection.Service) gorouter.Handler {
	return func(c gorouter.Context) error {
		var vars brigadeIDVars
		if err := c.Vars(&vars); err != nil {
			return fmt.Errorf("failed to read brigade id: %w", err)
		}

		var pageVars pagination.Pagination
		if err := c.Vars(&pageVars); err != nil {
			return fmt.Errorf("failed to read pagination: %w", err)
		}

		response, err := s.GetByBrigade(c.Ctx(), vars.BrigadeID, pageVars)
		if err != nil {
			return fmt.Errorf("failed to get inspections by brigade id: %w", err)
		}

		return c.WriteJson(http.StatusOK, response)
	}
}

type inspectionIDVars struct {
	ID int `path:"id"`
}

// AttachPhotoToInspection godoc
// @Summary Attach inspection photo
// @Description Uploads one device or seal photo and attaches it to an inspection.
// @Tags inspections
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Inspection ID"
// @Param Photo formData file true "Inspection photo"
// @Param AttachmentType formData int true "Attachment type: 1=device photo, 2=seal photo"
// @Param DeviceID formData int false "Device ID, required when AttachmentType is 1"
// @Param SealID formData int false "Seal ID, required when AttachmentType is 2"
// @Success 200 {object} inspection.Attachment
// @Failure 400 {object} gorouter.ErrorResponse
// @Failure 404 {object} gorouter.ErrorResponse
// @Failure 500 {object} gorouter.ErrorResponse
// @Router /inspections/{id}/photo [post]
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

// FinishInspection godoc
// @Summary Finish inspection
// @Description Saves inspection results, generated data, and completion state.
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path int true "Inspection ID"
// @Param request body inspection.FinishInspectionRequest true "Inspection completion payload"
// @Success 200 {object} inspection.Inspection
// @Failure 400 {object} gorouter.ErrorResponse
// @Failure 404 {object} gorouter.ErrorResponse
// @Failure 500 {object} gorouter.ErrorResponse
// @Router /inspections/{id}/finish [patch]
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
