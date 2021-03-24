package updater

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
)

const latestFileName = "latest.txt"

type AvailableUpdate struct {
	Asset    string
	Channel  string
	Version  string
	Specs    map[string]string
	FilePath string
}

/*
The way updates are looked for depends on the file structure given by the use case.

UpdateSource/AssetName/Channel/latest.txt pointing to the latest major
UpdateSource/AssetName/Channel/Major/latest.txt pointing to the latest minor or patch
UpdateSource/AssetName/Channel/Major/AssetName_Version_Specs_FileExtension the actual major´s minor or patch file

Updates have to be looked for depending on their major. Therefore one call of the getUpdatesInFolder function
is used to get current major´s minor or patch updater infos. A second call is used to look for major updater infos.
*/

// CheckForUpdates
// Looks for the latest updates available at the updates source. Returns information about the newest available major, minor and patch updates.
func (a Asset) CheckForUpdates() (availableUpdates []UpdateInfo, updateFound bool, err error) {
	currentMajor, _, _, err := getSemanticVersioningParts(a.AssetVersion)
	if err != nil {
		return nil, false, err
	}

	latestMajor, err := a.getLatestMajor()
	if err != nil {
		return nil, false, err
	}

	if latestMajor != currentMajor {
		majorUpdate, majorUpdateFound, err := a.getUpdatesInFolder(latestMajor)
		if err != nil {
			log.Println(err)
		}
		if majorUpdateFound {
			availableUpdates = append(availableUpdates, *majorUpdate)
			updateFound = true
		}
	}

	patchOrMinorUpdate, patchOrMinorUpdateFound, err := a.getUpdatesInFolder(currentMajor)
	if err != nil {
		log.Println(err)
	}
	if patchOrMinorUpdateFound {
		availableUpdates = append(availableUpdates, *patchOrMinorUpdate)
		updateFound = true
	}
	return availableUpdates, updateFound, nil
}

func (a Asset) getUpdatesInFolder(majorVersion string) (update *UpdateInfo, updateFound bool, err error) {
	latest, err := a.getLatestVersionInMajorDir(majorVersion)
	if err != nil {
		return nil, false, err
	}

	updateIsNewerThanCurrent, err := isUpdateNewerThanCurrent(a.AssetVersion, latest)
	if err != nil {
		return nil, false, err
	}
	if !updateIsNewerThanCurrent {
		return nil, false, nil
	}

	updatePath, err := a.getUpdatePathFromJson(majorVersion, latest)
	if err != nil {
		return nil, false, err
	}
	updateType, err := getUpdateType(a.AssetVersion, latest)
	if err != nil {
		return nil, false, err
	}

	return &UpdateInfo{
		Version: latest,
		Path:    updatePath,
		Type:    updateType,
	}, true, nil
}

func (a Asset) getLatestMajor() (latestMajor string, err error) {
	path := a.getPathToLatestMajor()
	data, err := a.Client.readData(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a Asset) getLatestVersionInMajorDir(major string) (version string, err error) {
	path := a.getPathToLatestPatchInMajorDir(major)
	data, err := a.Client.readData(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getUpdateType(currentVersion string, newVersion string) (updateType string, err error) {
	cMajor, cMinor, _, err := getSemanticVersioningParts(currentVersion)
	if err != nil {
		return "", err
	}

	nMajor, nMinor, _, err := getSemanticVersioningParts(newVersion)
	if err != nil {
		return "", err
	}

	if nMajor > cMajor {
		return "major", nil
	}
	if nMinor > cMinor {
		return "minor", nil
	}
	return "patch", nil
}

func (a Asset) getUpdatePathFromJson(majorVersion string, latestMinor string) (updatePath string, err error) {
	versionJsonPath := a.getPathToCdnVersionJson(majorVersion, latestMinor)
	data, err := a.Client.readData(versionJsonPath)
	var availableUpdates []AvailableUpdate
	if err = json.Unmarshal(data, &availableUpdates); err != nil {
		return "", err
	}
	for _, update := range availableUpdates {
		if matches := a.isUpdateValid(update, latestMinor); matches {
			return update.FilePath, nil
		}
	}
	return updatePath, errors.New("no matching update in version json at update server")
}

func (a Asset) isUpdateValid(availableUpdate AvailableUpdate, latest string) (match bool) {
	assetSpecs := make([]string, 0, len(a.Specs))
	for k := range a.Specs {
		assetSpecs = append(assetSpecs, k)
	}
	sort.Strings(assetSpecs)

	if a.AssetName != availableUpdate.Asset {
		return false
	}
	if a.Channel != availableUpdate.Channel {
		return false
	}
	if latest != availableUpdate.Version {
		return false
	}

	updateSpecs := make([]string, 0, len(availableUpdate.Specs))
	for k := range availableUpdate.Specs {
		updateSpecs = append(updateSpecs, k)
	}
	sort.Strings(updateSpecs)

	if len(assetSpecs) != len(updateSpecs) {
		return false
	}
	for i := range assetSpecs {
		if a.Specs[assetSpecs[i]] != availableUpdate.Specs[updateSpecs[i]] {
			return false
		}
	}
	return true
}
