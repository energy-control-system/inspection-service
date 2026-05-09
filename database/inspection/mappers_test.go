package inspection

import (
	"testing"
	"time"
)

func TestMapFromDBIncludesAttachments(t *testing.T) {
	createdAt := time.Date(2026, time.May, 9, 12, 0, 0, 0, time.UTC)

	got := MapFromDB(Inspection{
		ID:     10,
		TaskID: 20,
		Attachments: []Attachment{
			{
				ID:           1,
				InspectionID: 10,
				Type:         2,
				FileID:       30,
				CreatedAt:    createdAt,
			},
		},
	})

	if len(got.Attachments) != 1 {
		t.Fatalf("len(got.Attachments) = %d, want 1", len(got.Attachments))
	}

	attachment := got.Attachments[0]
	if attachment.ID != 1 {
		t.Fatalf("attachment.ID = %d, want 1", attachment.ID)
	}
	if attachment.InspectionID != 10 {
		t.Fatalf("attachment.InspectionID = %d, want 10", attachment.InspectionID)
	}
	if attachment.Type != 2 {
		t.Fatalf("attachment.Type = %d, want 2", attachment.Type)
	}
	if attachment.FileID != 30 {
		t.Fatalf("attachment.FileID = %d, want 30", attachment.FileID)
	}
	if !attachment.CreatedAt.Equal(createdAt) {
		t.Fatalf("attachment.CreatedAt = %s, want %s", attachment.CreatedAt, createdAt)
	}
}
