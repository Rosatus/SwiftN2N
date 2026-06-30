package edge

import "strings"

type LogClassification struct {
	State     State
	Message   string
	LastError string
	Matched   bool
	NoReply   bool
	Connected bool
}

func ClassifyLogLine(line string, noResponseCount int) LogClassification {
	lower := strings.ToLower(line)
	message := trimLogMessage(line)

	switch {
	case containsAny(lower, "permission denied", "authentication error", "auth error", "address already in use", "unable to open", "failed to add supernode"):
		return LogClassification{State: StateFailed, Message: message, LastError: message, Matched: true}
	case strings.Contains(lower, "no response from supernode") || strings.Contains(lower, "supernode not responding"):
		if noResponseCount >= 3 {
			return LogClassification{
				State:     StateFailed,
				Message:   "supernode did not respond; check port/protocol, key, community, and header encryption",
				LastError: "no response from supernode",
				Matched:   true,
				NoReply:   true,
			}
		}
		return LogClassification{
			State:   StateConnecting,
			Message: "waiting for supernode response; if UDP is blocked, try TCP mode",
			Matched: true,
			NoReply: true,
		}
	case containsAny(lower,
		"edge connected to supernode",
		"registered with supernode",
		"rx register_super_ack",
		"[ok] edge",
		"rx pong from supernode",
		"supernode responding",
	) || (strings.Contains(lower, "peer") && strings.Contains(lower, "brought up")):
		return LogClassification{State: StateConnected, Message: "edge connected", Matched: true, Connected: true}
	case containsAny(lower, "connecting", "register", "trying", "created local tap", "created tun", "successfully joined multicast"):
		return LogClassification{State: StateConnecting, Message: message, Matched: true}
	default:
		return LogClassification{}
	}
}

func RedactSecrets(value string, secrets []string) string {
	redacted := redactFlagValues(value)
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret != "" {
			redacted = strings.ReplaceAll(redacted, secret, "********")
		}
	}
	return redacted
}

func RedactArgs(args []string) []string {
	clone := append([]string(nil), args...)
	for i := 0; i < len(clone)-1; i++ {
		switch clone[i] {
		case "-k", "-J":
			clone[i+1] = "********"
		}
	}
	return clone
}

func redactFlagValues(value string) string {
	parts := strings.Fields(value)
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "-k", "-J":
			parts[i+1] = "********"
		}
	}
	return strings.Join(parts, " ")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func trimLogMessage(line string) string {
	line = strings.TrimSpace(line)
	if len(line) > 180 {
		return line[:180]
	}
	return line
}
