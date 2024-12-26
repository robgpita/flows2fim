package utils

import "os/exec"

// Check GDAL tools available checks if gdalbuildvrt is available in the environment.
func CheckGDALToolAvailable(tool string) bool {
	cmd := exec.Command(tool, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
