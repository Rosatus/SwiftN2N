//go:build !windows

package platform

import "testing"

func TestPrivilegedHelperPreservesCustomEdgeOverride(t *testing.T) {
	for _, key := range privilegedHelperEnvKeys() {
		if key == "SWIFTN2N_ALLOW_CUSTOM_EDGE_PATH" {
			return
		}
	}
	t.Fatalf("privileged helper must preserve SWIFTN2N_ALLOW_CUSTOM_EDGE_PATH for development overrides")
}
