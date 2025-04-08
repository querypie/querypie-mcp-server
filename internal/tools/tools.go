package tools

import (
	"os"
	"path/filepath"
	"strings"
)

func IsDev() bool {
	_, withGoRun := inspectRuntime()
	return withGoRun
}

func inspectRuntime() (baseDir string, withGoRun bool) {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// for running with go run
		withGoRun = true
		baseDir, _ = os.Getwd()
	} else if strings.Contains(os.Args[0], "/tmp/GoLand/") {
		// for GoLand
		withGoRun = true
		baseDir, _ = os.Getwd()
	} else {
		withGoRun = false
		baseDir = filepath.Dir(os.Args[0])
	}
	return
}
