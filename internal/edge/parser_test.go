package edge

import "testing"

func TestClassifyLogLine(t *testing.T) {
	tests := []struct {
		name            string
		line            string
		noResponseCount int
		wantMatched     bool
		wantState       State
		wantNoReply     bool
		wantConnected   bool
		wantMessage     string
	}{
		{
			name:        "connecting",
			line:        "registering with supernode",
			wantMatched: true,
			wantState:   StateConnecting,
		},
		{
			name:          "connected",
			line:          "edge connected to supernode",
			wantMatched:   true,
			wantState:     StateConnected,
			wantConnected: true,
		},
		{
			name:          "register super ack",
			line:          "Rx REGISTER_SUPER_ACK from C6:17:A2:42:27:B6 [203.0.113.10:9077]",
			wantMatched:   true,
			wantState:     StateConnected,
			wantConnected: true,
		},
		{
			name:          "ok banner",
			line:          "[OK] edge <<< ================ >>> supernode",
			wantMatched:   true,
			wantState:     StateConnected,
			wantConnected: true,
		},
		{
			name:          "pong",
			line:          "Rx PONG from supernode C6:17:A2:42:27:B6",
			wantMatched:   true,
			wantState:     StateConnected,
			wantConnected: true,
		},
		{
			name:        "first no response",
			line:        "supernode not responding",
			wantMatched: true,
			wantState:   StateConnecting,
			wantNoReply: true,
			wantMessage: "waiting for supernode response; if UDP is blocked, try TCP mode",
		},
		{
			name:            "third no response",
			line:            "no response from supernode",
			noResponseCount: 3,
			wantMatched:     true,
			wantState:       StateFailed,
			wantNoReply:     true,
			wantMessage:     "supernode did not respond; check port/protocol, key, community, and header encryption",
		},
		{
			name:        "auth failure",
			line:        "authentication error",
			wantMatched: true,
			wantState:   StateFailed,
		},
		{
			name:        "unrelated",
			line:        "trace packet 12",
			wantMatched: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ClassifyLogLine(test.line, test.noResponseCount)
			if got.Matched != test.wantMatched {
				t.Fatalf("Matched = %v, want %v", got.Matched, test.wantMatched)
			}
			if got.State != test.wantState {
				t.Fatalf("State = %q, want %q", got.State, test.wantState)
			}
			if got.NoReply != test.wantNoReply {
				t.Fatalf("NoReply = %v, want %v", got.NoReply, test.wantNoReply)
			}
			if got.Connected != test.wantConnected {
				t.Fatalf("Connected = %v, want %v", got.Connected, test.wantConnected)
			}
			if test.wantMessage != "" && got.Message != test.wantMessage {
				t.Fatalf("Message = %q, want %q", got.Message, test.wantMessage)
			}
		})
	}
}
