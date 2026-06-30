//go:build windows

package edge

import (
	"fmt"
	"io"
)

func RunStopHelper(pid int, stdout io.Writer, stderr io.Writer) int {
	_, _ = fmt.Fprintln(stderr, "Windows uses the application administrator manifest and taskkill; helper stop is not supported")
	return 2
}
