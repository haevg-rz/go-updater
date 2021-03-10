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
		return "", "", "", errors.New("invalid Version")
	}
	major = parts[0]
	minor = parts[1]
	patch = parts[2]
	return major, minor, patch, nil
}

func isUpdateNewerThanCurrent(currentVersion string, updateVersion string) (updateIsNewer bool, err error) {
	currentMajor, currentMinor, currentPatch, err := getSemanticVersioningParts(currentVersion)
	if err != nil {
		return false, err
	}

	updateMajor, updateMinor, updatePatch, err := getSemanticVersioningParts(updateVersion)
	if err != nil {
		return false, err
	}

	if updateMajor > currentMajor {
		return true, nil
	}

	if updateMajor == currentMajor && updateMinor > currentMinor {
		return true, nil
	}

	if updateMajor == currentMajor && updateMinor == currentMinor && updatePatch > currentPatch {
		return true, nil
	}
	return false, nil
}
