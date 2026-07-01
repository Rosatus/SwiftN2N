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

type PathResolution struct {
	Path   string
	Custom bool
}

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

func ResolveConfiguredPath(config Config) (PathResolution, error) {
	if config.AllowCustomEdgePath && strings.TrimSpace(config.EdgePath) != "" {
		path, err := ResolvePath(config.EdgePath)
		return PathResolution{Path: path, Custom: true}, err
	}
	if path, err := ResolveBundledPath(); err == nil {
		return PathResolution{Path: path}, nil
	}
	path, err := ResolvePath("")
	return PathResolution{Path: path, Custom: true}, err
}

func ResolveBundledPath() (string, error) {
	name := platform.EdgeExecutableName()
	exePath := ""
	if exe, err := os.Executable(); err == nil {
		exePath = exe
	}
	cwd := ""
	if currentDir, err := os.Getwd(); err == nil {
		cwd = currentDir
	}
	for _, candidate := range bundledCandidates(exePath, cwd, name) {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("bundled edge binary not found")
}

func bundledCandidates(exePath string, cwd string, name string) []string {
	candidates := make([]string, 0, 6)
	if exePath != "" {
		dir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(dir, "bin", runtime.GOOS, runtime.GOARCH, name),
			filepath.Join(dir, "bin", runtime.GOOS, name),
			filepath.Join(dir, name),
		)
	}
	if cwd != "" {
		candidates = append(candidates,
			filepath.Join(cwd, "bin", runtime.GOOS, runtime.GOARCH, name),
			filepath.Join(cwd, "bin", runtime.GOOS, name),
			filepath.Join(cwd, name),
		)
	}
	return candidates
}

func CustomEdgePathAllowedForHelper() bool {
	return os.Getenv("SWIFTN2N_ALLOW_CUSTOM_EDGE_PATH") == "1"
}
