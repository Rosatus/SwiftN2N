package edge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"SwiftN2N/internal/platform"
)

func ResolvePath(explicit string) (string, error) {
	candidates := make([]string, 0, 8)
	if strings.TrimSpace(explicit) != "" {
		candidates = append(candidates, strings.TrimSpace(explicit))
	}
	if envPath := strings.TrimSpace(os.Getenv("SWIFTN2N_EDGE_PATH")); envPath != "" {
		candidates = append(candidates, envPath)
	}

	name := platform.EdgeExecutableName()
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "bin", runtime.GOOS, runtime.GOARCH, name),
			filepath.Join(cwd, "bin", runtime.GOOS, name),
			filepath.Join(cwd, name),
		)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "bin", runtime.GOOS, runtime.GOARCH, name),
			filepath.Join(dir, "bin", runtime.GOOS, name),
			filepath.Join(dir, name),
		)
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("edge binary not found; set edgePath or SWIFTN2N_EDGE_PATH")
}
