package edge

import (
	"errors"
	"fmt"
	"net/netip"
	"runtime"
	"strconv"
	"strings"

	"SwiftN2N/internal/platform"
)

type CommandSpec struct {
	Args    []string
	Env     []string
	Secrets []string
}

func BuildCommand(config Config) (CommandSpec, error) {
	var spec CommandSpec

	community := strings.TrimSpace(config.Community)
	if community == "" {
		return spec, errors.New("community is required")
	}
	if hasControlChars(community) {
		return spec, errors.New("community contains invalid control characters")
	}

	supernodes := CleanList(config.Supernodes)
	if len(supernodes) == 0 {
		return spec, errors.New("at least one supernode is required")
	}

	spec.Args = append(spec.Args, "-c", community)
	for _, supernode := range supernodes {
		if hasUnsafeCLIValue(supernode) {
			return spec, fmt.Errorf("supernode %q contains invalid whitespace or control characters", supernode)
		}
		spec.Args = append(spec.Args, "-l", supernode)
	}

	if value := strings.TrimSpace(config.Address); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("address contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-a", value)
	}
	if value := strings.TrimSpace(config.Key); value != "" {
		spec.Env = append(spec.Env, "N2N_KEY="+value)
		spec.Secrets = append(spec.Secrets, value)
	}
	flag, err := CipherFlag(config.Cipher)
	if err != nil {
		return spec, err
	}
	if flag != "" {
		spec.Args = append(spec.Args, flag)
	}
	if config.HeaderEncryption {
		spec.Args = append(spec.Args, "-H")
	}
	if value := strings.TrimSpace(config.MAC); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("MAC address contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-m", value)
	}
	if value := strings.TrimSpace(config.DeviceName); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("device name contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-d", value)
	}
	if config.MTU < 0 {
		return spec, errors.New("MTU cannot be negative")
	}
	if config.MTU > 0 {
		spec.Args = append(spec.Args, "-M", strconv.Itoa(config.MTU))
	}
	if value := strings.TrimSpace(config.LocalPort); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("local port contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-p", value)
	}
	if config.ManagementPort < 0 || config.ManagementPort > 65535 {
		return spec, errors.New("management port must be between 0 and 65535")
	}
	if config.ManagementPort > 0 {
		spec.Args = append(spec.Args, "-t", strconv.Itoa(config.ManagementPort))
	}
	if value := strings.TrimSpace(config.AuthUsername); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("auth username contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-I", value)
	}
	if value := strings.TrimSpace(config.AuthPassword); value != "" {
		spec.Env = append(spec.Env, "N2N_PASSWORD="+value)
		spec.Secrets = append(spec.Secrets, value)
	}
	if value := strings.TrimSpace(config.FederationPublicKey); value != "" {
		if hasUnsafeCLIValue(value) {
			return spec, errors.New("federation public key contains invalid whitespace or control characters")
		}
		spec.Args = append(spec.Args, "-P", value)
	}

	switch strings.ToLower(strings.TrimSpace(config.SupernodeOnly)) {
	case "udp":
		spec.Args = append(spec.Args, "-S1")
	case "tcp":
		spec.Args = append(spec.Args, "-S2")
	case "":
	default:
		return spec, errors.New("supernodeOnly must be empty, udp, or tcp")
	}

	switch strings.ToLower(strings.TrimSpace(config.Compression)) {
	case "lzo":
		spec.Args = append(spec.Args, "-z1")
	case "zstd":
		spec.Args = append(spec.Args, "-z2")
	case "":
	default:
		return spec, errors.New("compression must be empty, lzo, or zstd")
	}

	if config.AcceptMulticast {
		spec.Args = append(spec.Args, "-E")
	}
	if config.EnableRouting {
		spec.Args = append(spec.Args, "-r")
	}
	for _, route := range CleanList(config.Routes) {
		if err := validateRoute(route); err != nil {
			return spec, err
		}
		spec.Args = append(spec.Args, "-n", route)
	}
	for _, rule := range CleanList(config.TrafficRules) {
		if hasControlChars(rule) {
			return spec, errors.New("traffic rule contains invalid control characters")
		}
		spec.Args = append(spec.Args, "-R", rule)
	}
	if config.WindowsMetric < 0 {
		return spec, errors.New("Windows metric cannot be negative")
	}
	if runtime.GOOS == "windows" && config.WindowsMetric > 0 {
		spec.Args = append(spec.Args, "-x", strconv.Itoa(config.WindowsMetric))
	}
	if platform.SupportsForegroundFlag() {
		spec.Args = append(spec.Args, "-f")
	}
	if config.Verbose < 0 || config.Verbose > 5 {
		return spec, errors.New("verbose level must be between 0 and 5")
	}
	for i := 0; i < config.Verbose; i++ {
		spec.Args = append(spec.Args, "-v")
	}

	return spec, nil
}

func CipherFlag(cipher string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(cipher)) {
	case "":
		return "", nil
	case "none":
		return "-A1", nil
	case "twofish":
		return "-A2", nil
	case "aes":
		return "-A3", nil
	case "chacha20", "chacha":
		return "-A4", nil
	case "speck":
		return "-A5", nil
	default:
		return "", fmt.Errorf("unsupported cipher %q", cipher)
	}
}

func CleanList(values []string) []string {
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

func validateRoute(route string) error {
	if hasControlChars(route) {
		return errors.New("route contains invalid control characters")
	}
	parts := strings.Split(route, ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("route %q must use network/prefix:gateway format", route)
	}
	if _, err := netip.ParsePrefix(parts[0]); err != nil {
		return fmt.Errorf("route %q has invalid network prefix", route)
	}
	if _, err := netip.ParseAddr(parts[1]); err != nil {
		return fmt.Errorf("route %q has invalid gateway address", route)
	}
	return nil
}

func hasUnsafeCLIValue(value string) bool {
	return strings.ContainsAny(value, " \t\r\n\x00")
}

func hasControlChars(value string) bool {
	return strings.ContainsAny(value, "\r\n\x00")
}
