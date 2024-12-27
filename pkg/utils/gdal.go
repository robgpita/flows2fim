package utils

import (
	"os/exec"
	"strings"
)

// Check GDAL tools available checks if gdalbuildvrt is available in the environment.
func CheckGDALToolAvailable(tool string) bool {
	cmd := exec.Command(tool, "--version")
	if out, err := cmd.Output(); err != nil {
		outStr := string(out)
		return strings.Contains(strings.ToLower(outStr), "released") // https://github.com/OSGeo/gdal/issues/11550
	}
	return true
}
