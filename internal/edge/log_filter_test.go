package edge

import "testing"

func TestNoisyLogLabel(t *testing.T) {
	tests := []struct {
		line      string
		wantOK    bool
		wantLabel string
	}{
		{
			line:      "30/Jun/2026 17:32:07 [edge_utils.c:2340] dropping Tx multicast",
			wantOK:    true,
			wantLabel: "dropping Tx multicast",
		},
		{
			line:      "DROP packet before first registration with supernode",
			wantOK:    true,
			wantLabel: "DROP packet before first registration with supernode",
		},
		{
			line:      "Rx TAP packet ( 128) for 01:00:5E:00:00:FB",
			wantOK:    true,
			wantLabel: "Rx TAP packet",
		},
		{
			line:   "Rx REGISTER_SUPER_ACK from C6:17:A2:42:27:B6",
			wantOK: false,
		},
	}

	for _, test := range tests {
		t.Run(test.line, func(t *testing.T) {
			gotLabel, gotOK := noisyLogLabel(test.line)
			if gotOK != test.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, test.wantOK)
			}
			if gotLabel != test.wantLabel {
				t.Fatalf("label = %q, want %q", gotLabel, test.wantLabel)
			}
		})
	}
}
