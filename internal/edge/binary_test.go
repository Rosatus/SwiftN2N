package edge

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"SwiftN2N/internal/platform"
)

func TestResolveConfiguredPathUsesBundledByDefault(t *testing.T) {
	tempDir := t.TempDir()
	bundled := filepath.Join(tempDir, "bin", runtime.GOOS, runtime.GOARCH, platform.EdgeExecutableName())
	if err := os.MkdirAll(filepath.Dir(bundled), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(bundled, []byte("edge"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	custom := filepath.Join(tempDir, "custom-edge")
	if runtime.GOOS == "windows" {
		custom += ".exe"
	}
	if err := os.WriteFile(custom, []byte("custom"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatalf("Chdir cleanup returned error: %v", err)
		}
	})
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}

	resolved, err := ResolveConfiguredPath(Config{EdgePath: custom})
	if err != nil {
		t.Fatalf("ResolveConfiguredPath returned error: %v", err)
	}
	if resolved.Custom {
		t.Fatalf("resolved path should not be custom by default: %+v", resolved)
	}
	if resolved.Path != bundled {
		t.Fatalf("Path = %q, want %q", resolved.Path, bundled)
	}

	resolved, err = ResolveConfiguredPath(Config{EdgePath: custom, AllowCustomEdgePath: true})
	if err != nil {
		t.Fatalf("ResolveConfiguredPath with override returned error: %v", err)
	}
	if !resolved.Custom {
		t.Fatalf("resolved path should be custom with explicit override: %+v", resolved)
	}
	if resolved.Path != custom {
		t.Fatalf("Path = %q, want %q", resolved.Path, custom)
	}
}

func TestBundledCandidatesPreferExecutableDirectory(t *testing.T) {
	name := platform.EdgeExecutableName()
	candidates := bundledCandidates(
		filepath.Join("app", "SwiftN2N"),
		filepath.Join("cwd"),
		name,
	)

	if len(candidates) < 4 {
		t.Fatalf("expected executable and cwd candidates, got %v", candidates)
	}
	want := filepath.Join("app", "bin", runtime.GOOS, runtime.GOARCH, name)
	if candidates[0] != want {
		t.Fatalf("first candidate = %q, want %q", candidates[0], want)
	}
	cwdCandidate := filepath.Join("cwd", "bin", runtime.GOOS, runtime.GOARCH, name)
	for index, candidate := range candidates {
		if candidate == cwdCandidate && index < 3 {
			t.Fatalf("cwd candidate %q was checked before executable directory candidates: %v", candidate, candidates)
		}
	}
}
