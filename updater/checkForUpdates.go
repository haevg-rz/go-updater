package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
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
func (asset Asset) CheckForUpdates() (availableUpdates []UpdateInfo, err error) {
	currentMajor, _, _, err := getSemanticVersioningParts(asset.AssetVersion)
	if err != nil {
		return availableUpdates, err
	}

	//TODO asset hat client und client hat Methode getLatestMajor asset.UpdateClient.GetLatestMajor
	latestMajor := asset.getLatestMajor()
	if latestMajor != currentMajor {
		//Todo return updater found bool
		majorUpdate := asset.getUpdatesInFolder(latestMajor)
		if majorUpdate != nil {
			availableUpdates = append(availableUpdates, *majorUpdate)
		}
	}

	patchOrMinorUpdate := asset.getUpdatesInFolder(currentMajor)
	if patchOrMinorUpdate != nil {
		availableUpdates = append(availableUpdates, *patchOrMinorUpdate)
	}
	return availableUpdates, nil
}

func (asset Asset) getUpdatesInFolder(majorVersion string) (update *UpdateInfo) {
	latest, err := asset.getLatest(majorVersion)
	if err != nil {
		return
	}

	if isUpdateNewerThanCurrent(asset.AssetVersion, latest) {
		updatePath, err := asset.getUpdatePathFromJson(majorVersion, latest)
		if err != nil {
			return
		}
		return &UpdateInfo{
			Version: latest,
			Path:    updatePath,
			Type:    getUpdateType(asset.AssetVersion, latest),
		}
	}
	return
}

func (asset Asset) getLatestMajor() (latestMajor string) {
	path := filepath.Join(asset.AssetName, asset.Channel, latestFileName)
	data, err := asset.Client.readData(path)
	printErrors(err)

	latestMajor = string(data)
	printErrors(err)
	return
}

func (asset Asset) getLatest(major string) (version string, err error) {
	majorPath := asset.createMajorPath(major)
	path := filepath.Join(majorPath, latestFileName)
	data, err := asset.Client.readData(path)
	return string(data), err
}

func getUpdateType(currentVersion string, newVersion string) (semVerPart string) {
	cMajor, cMinor, _, err := getSemanticVersioningParts(currentVersion)
	printErrors(err)
	nMajor, nMinor, _, err := getSemanticVersioningParts(newVersion)
	printErrors(err)

	if nMajor > cMajor {
		return "major"
	}
	if nMinor > cMinor {
		return "minor"
	}
	return "patch"
}

func (asset Asset) getUpdatePathFromJson(majorVersion string, latestMinor string) (updatePath string, err error) {
	majorPath := asset.createMajorPath(majorVersion)
	jsonPath := filepath.Join(majorPath, fmt.Sprint(latestMinor, ".json"))
	data, err := asset.Client.readData(jsonPath)
	//TODO Slice direkt im JSON
	var jsonContent struct {
		AvailableUpdates []AvailableUpdate
	}
	if err = json.Unmarshal(data, &jsonContent); err != nil {
		return
	}
	for _, update := range jsonContent.AvailableUpdates {
		if matches := asset.isUpdateValidForAsset(update, latestMinor); matches {
			return update.FilePath, err
		}
	}
	return updatePath, errors.New("no matching updater in version json at updateServer")
}

func (asset Asset) isUpdateValidForAsset(availableUpdate AvailableUpdate, latest string) (match bool) {
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
