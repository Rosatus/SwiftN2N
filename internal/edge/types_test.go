package edge

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnvironmentStatusMissingMarshalsAsArray(t *testing.T) {
	payload, err := json.Marshal(EnvironmentStatus{Missing: []string{}})
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if !strings.Contains(string(payload), `"missing":[]`) {
		t.Fatalf("missing should marshal as an empty array, got %s", payload)
	}
}
