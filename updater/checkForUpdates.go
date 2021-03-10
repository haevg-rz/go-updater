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
func (asset Asset) CheckForUpdates() (availableUpdates []UpdateInfo, updateFound bool, err error) {
	currentMajor, _, _, err := getSemanticVersioningParts(asset.AssetVersion)
	if err != nil {
		return nil, false, err
	}

	latestMajor, err := asset.getLatestMajor()
	if err != nil {
		return nil, false, err
	}

	if latestMajor != currentMajor {
		majorUpdate, majorUpdateFound, err := asset.getUpdatesInFolder(latestMajor)
		if err != nil {
			log.Println(err)
		}
		if majorUpdateFound == true {
			availableUpdates = append(availableUpdates, *majorUpdate)
			updateFound = true
		}
	}

	patchOrMinorUpdate, patchOrMinorUpdateFound, err := asset.getUpdatesInFolder(currentMajor)
	if err != nil {
		log.Println(err)
	}
	if patchOrMinorUpdateFound == true {
		availableUpdates = append(availableUpdates, *patchOrMinorUpdate)
		updateFound = true
	}
	return availableUpdates, updateFound, nil
}

func (asset Asset) getUpdatesInFolder(majorVersion string) (update *UpdateInfo, updateFound bool, err error) {
	latest, err := asset.getLatestVersionInMajorDir(majorVersion)
	if err != nil {
		return nil, false, err
	}

	updateIsNewerThanCurrent, err := isUpdateNewerThanCurrent(asset.AssetVersion, latest)
	if err != nil {
		return nil, false, err
	}
	if !updateIsNewerThanCurrent {
		return nil, false, nil
	}

	updatePath, err := asset.getUpdatePathFromJson(majorVersion, latest)
	if err != nil {
		return nil, false, err
	}
	updateType, err := getUpdateType(asset.AssetVersion, latest)
	if err != nil {
		return nil, false, err
	}

	return &UpdateInfo{
		Version: latest,
		Path:    updatePath,
		Type:    updateType,
	}, true, nil
}

func (asset Asset) getLatestMajor() (latestMajor string, err error) {
	path := asset.getPathToLatestMajor()
	data, err := asset.Client.readData(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (asset Asset) getLatestVersionInMajorDir(major string) (version string, err error) {
	path := asset.getPathToLatestPatchInMajorDir(major)
	data, err := asset.Client.readData(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getUpdateType(currentVersion string, newVersion string) (semVerPart string, err error) {
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

func (asset Asset) getUpdatePathFromJson(majorVersion string, latestMinor string) (updatePath string, err error) {
	versionJsonPath := asset.getPathToCdnVersionJson(majorVersion, latestMinor)
	data, err := asset.Client.readData(versionJsonPath)
	var availableUpdates []AvailableUpdate
	if err = json.Unmarshal(data, &availableUpdates); err != nil {
		return "", err
	}
	for _, update := range availableUpdates {
		if matches := asset.isUpdateValid(update, latestMinor); matches {
			return update.FilePath, nil
		}
	}
	return updatePath, errors.New("no matching update in version json at update server")
}

func (asset Asset) isUpdateValid(availableUpdate AvailableUpdate, latest string) (match bool) {
	assetSpecs := make([]string, 0, len(asset.Specs))
	for k := range asset.Specs {
		assetSpecs = append(assetSpecs, k)
	}
	sort.Strings(assetSpecs)

	if asset.AssetName != availableUpdate.Asset {
		return false
	}
	if asset.Channel != availableUpdate.Channel {
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
		if asset.Specs[assetSpecs[i]] != availableUpdate.Specs[updateSpecs[i]] {
			return false
		}
	}
	return true
}
