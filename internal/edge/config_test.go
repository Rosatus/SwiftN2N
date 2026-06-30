package edge

import (
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestBuildCommandBasicConfig(t *testing.T) {
	spec, err := BuildCommand(Config{
		Supernodes:       []string{"sn.example.net:7777"},
		Community:        "office",
		Address:          "10.10.10.12",
		Key:              "shared-secret",
		Cipher:           "chacha20",
		HeaderEncryption: true,
		MTU:              1290,
		Verbose:          2,
	})
	if err != nil {
		t.Fatalf("BuildCommand returned error: %v", err)
	}

	wantArgs := []string{"-c", "office", "-l", "sn.example.net:7777", "-a", "10.10.10.12", "-A4", "-H", "-M", "1290"}
	for _, want := range wantArgs {
		if !slices.Contains(spec.Args, want) {
			t.Fatalf("args %v missing %q", spec.Args, want)
		}
	}
	if runtime.GOOS != "windows" && !slices.Contains(spec.Args, "-f") {
		t.Fatalf("unix args %v missing -f", spec.Args)
	}
	if got := strings.Join(spec.Args, " "); strings.Contains(got, "shared-secret") {
		t.Fatalf("secret leaked into args: %s", got)
	}
	if !slices.Contains(spec.Env, "N2N_KEY=shared-secret") {
		t.Fatalf("env %v missing N2N_KEY", spec.Env)
	}
}

func TestBuildCommandAdvancedParams(t *testing.T) {
	spec, err := BuildCommand(Config{
		Supernodes:          []string{"sn.example.net:7777", "backup.example.net:7777"},
		Community:           "office",
		Cipher:              "aes",
		MAC:                 "DE:AD:BE:EF:10:12",
		DeviceName:          "swiftn2n0",
		LocalPort:           "0.0.0.0:7777",
		ManagementPort:      5644,
		AuthUsername:        "edge-01",
		AuthPassword:        "auth-secret",
		FederationPublicKey: "public-key",
		SupernodeOnly:       "tcp",
		Compression:         "zstd",
		AcceptMulticast:     true,
		EnableRouting:       true,
		Routes:              []string{"10.42.0.0/16:10.10.10.1"},
		TrafficRules:        []string{"accept ip"},
		Verbose:             1,
	})
	if err != nil {
		t.Fatalf("BuildCommand returned error: %v", err)
	}

	joined := strings.Join(spec.Args, " ")
	for _, want := range []string{"-A3", "-m DE:AD:BE:EF:10:12", "-d swiftn2n0", "-p 0.0.0.0:7777", "-t 5644", "-I edge-01", "-P public-key", "-S2", "-z2", "-E", "-r", "-n 10.42.0.0/16:10.10.10.1", "-R accept ip"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args %q missing %q", joined, want)
		}
	}
	if strings.Contains(joined, "auth-secret") {
		t.Fatalf("auth password leaked into args: %s", joined)
	}
	if !slices.Contains(spec.Env, "N2N_PASSWORD=auth-secret") {
		t.Fatalf("env %v missing N2N_PASSWORD", spec.Env)
	}
}

func TestBuildCommandRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{name: "missing community", config: Config{Supernodes: []string{"sn:7777"}}},
		{name: "missing supernode", config: Config{Community: "office"}},
		{name: "bad cipher", config: Config{Community: "office", Supernodes: []string{"sn:7777"}, Cipher: "rot13"}},
		{name: "bad route", config: Config{Community: "office", Supernodes: []string{"sn:7777"}, Routes: []string{"10.42.0.0/16"}}},
		{name: "bad management port", config: Config{Community: "office", Supernodes: []string{"sn:7777"}, ManagementPort: 70000}},
		{name: "bad verbose", config: Config{Community: "office", Supernodes: []string{"sn:7777"}, Verbose: 6}},
		{name: "supernode whitespace", config: Config{Community: "office", Supernodes: []string{"sn example:7777"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := BuildCommand(test.config); err == nil {
				t.Fatalf("BuildCommand returned nil error")
			}
		})
	}
}

func TestRedactSecrets(t *testing.T) {
	got := RedactSecrets("edge -k shared-secret -J auth-secret extra shared-secret", []string{"shared-secret", "auth-secret"})
	if strings.Contains(got, "shared-secret") || strings.Contains(got, "auth-secret") {
		t.Fatalf("secrets not redacted: %s", got)
	}
	if strings.Count(got, "********") < 3 {
		t.Fatalf("expected redaction markers in %q", got)
	}
}
