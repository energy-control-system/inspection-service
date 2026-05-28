package handler

import (
	"inspection-service/service/inspection"
	"testing"

	routerreflect "github.com/sunshineOfficial/golib/gohttp/gorouter/reflect"
)

func TestInspectionListQueryVarsReadsPaginationWithSort(t *testing.T) {
	var vars inspectionListQueryVars
	err := routerreflect.SetValuesToItem(map[string][]string{
		"limit":  {"10"},
		"offset": {"20"},
		"sort":   {"asc"},
	}, "query", &vars)
	if err != nil {
		t.Fatalf("SetValuesToItem returned error: %v", err)
	}

	if vars.Limit != 10 {
		t.Fatalf("limit = %d, want 10", vars.Limit)
	}
	if vars.Offset != 20 {
		t.Fatalf("offset = %d, want 20", vars.Offset)
	}
	if vars.Sort != inspection.SortAsc {
		t.Fatalf("sort = %q, want %q", vars.Sort, inspection.SortAsc)
	}
}
