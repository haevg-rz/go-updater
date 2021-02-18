package updater

import (
	"errors"
	"strings"
)

func getSemanticVersioningParts(version string) (major string, minor string, patch string, err error) {
	major = "0"
	minor = "0"
	patch = "0"
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		err = errors.New("invalid Version")
		return
	}
	major = parts[0]
	minor = parts[1]
	patch = parts[2]
	return
}

func isUpdateNewerThanCurrent(currentVersion string, updateVersion string) bool {

	if strings.Contains(strings.ToLower(currentVersion), "dev") {
		return true
	}

	currentMajor, currentMinor, currentPatch, err := getSemanticVersioningParts(currentVersion)
	if err != nil {
		return false
	}

	updateMajor, updateMinor, updatePatch, err := getSemanticVersioningParts(updateVersion)
	if err != nil {
		return false
	}

	if updateMajor > currentMajor {
		return true
	}

	if updateMajor == currentMajor && updateMinor > currentMinor {
		return true
	}

	if updateMajor == currentMajor && updateMinor == currentMinor && updatePatch > currentPatch {
		return true
	}
	return false
}
