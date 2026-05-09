package inspection

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestInspectedDeviceJSONUsesPascalCaseFields(t *testing.T) {
	device := InspectedDevice{
		ID:           900002,
		DeviceID:     920081,
		InspectionID: 900002,
		Value:        decimal.RequireFromString("4605.96"),
		Consumption:  decimal.RequireFromString("122.96"),
		CreatedAt:    time.Date(2025, time.January, 1, 7, 27, 0, 0, time.UTC),
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	var got map[string]any
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	for _, key := range []string{"ID", "DeviceID", "InspectionID", "Value", "Consumption", "CreatedAt"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("missing key %q in %s", key, data)
		}
	}

	for _, key := range []string{"id", "device_id", "inspection_id", "value", "consumption", "created_at"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unexpected key %q in %s", key, data)
		}
	}
}
