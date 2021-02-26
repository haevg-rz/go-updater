//Welcome to the updater Module

//This file contains the exposed Functions. If you are new to the project and want to understand the code,
//its recommended to start here!

package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type Asset struct {
	AssetName     string
	AssetVersion  string
	Channel       string
	Client        Client
	CdnBaseUrl    string
	DoMajorUpdate bool
	Specs         map[string]string
	TargetFolder  string
}

type UpdateInfo struct {
	Version string
	Path    string
	Type    string
}

// SelfUpdate
// Looks for the latest available updates. Applies the newest updater, terminating the running process and exchanging the executable files. Then restarts the application.
func (asset Asset) SelfUpdate() (updatedTo UpdateInfo, updated bool, err error) {
	availableUpdates, updateFound, err := asset.CheckForUpdates()
	if err != nil {
		return updatedTo, false, err
	}
	if !updateFound {
		return updatedTo, false, nil
	}
	latestUpdate := asset.getLatestAllowedUpdate(availableUpdates)
	err = asset.importSelfUpdate(latestUpdate.Path)
	if err != nil {
		return updatedTo, false, err
	}
	err = asset.applySelfUpdate()
	if err != nil {
		return updatedTo, false, err
	}
	updatedTo = latestUpdate
	return updatedTo, true, nil
}

// Update
// Looks for the latest available updates of an external Asset. Applies the newest updater and writes a versionJson into the asset folder, which points to the new version.
func (asset Asset) Update() (updatedTo UpdateInfo, updated bool, err error) {
	availableUpdates, updateFound, err := asset.CheckForUpdates()
	if err != nil {
		return
	}

	if !updateFound {
		return updatedTo, false, nil
	}

	latestUpdate := asset.getLatestAllowedUpdate(availableUpdates)
	err = asset.importUpdate(latestUpdate.Path)
	if err != nil {
		return latestUpdate, false, err
	}
	updatedTo = latestUpdate
	asset.AssetVersion = updatedTo.Version
	err = asset.writeVersionJson(latestUpdate.Version)
	if err != nil {
		return latestUpdate, false, err
	}
	return latestUpdate, true, err
}

func (asset Asset) getLatestAllowedUpdate(availableUpdates []UpdateInfo) (update UpdateInfo) {
	if asset.DoMajorUpdate {
		for _, update = range availableUpdates {
			if update.Type == "major" {
				return
			}
		}
	}
	for _, update = range availableUpdates {
		if update.Type == "minor" {
			return
		}
	}
	for _, update = range availableUpdates {
		if update.Type == "patch" {
			return
		}
	}
	return
}

//GetVersion
//Get a semantic Versioning string for the asset that is to be updated. Looks for a Version Json, written every time the asset
//is updated by this module. If the json can not be found, a default version of 0.0.0 is returned.
func GetVersion(targetFolder string, assetName string) (currentVersion string) {
	const defaultVersion = "0.0.0"
	currentVersion = defaultVersion
	fileName := fmt.Sprint(assetName, "_Version.Json")
	path := filepath.Join(targetFolder, fileName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	var content struct {
		Version string
	}
	if err = json.Unmarshal(data, &content); err != nil {
		return
	}
	return content.Version
}

//Background
//Starts looking for updates in a specified interval. Use the allowUpdate function to enable/disable updates.
func (asset Asset) Background(interval time.Duration, executeUpdateCallback func() (bool, error), skipUpdate func() bool, executeAfterUpdateCallback func() error) (err error) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				if skipUpdate() {
					log.Println("update skipped at,", t)
					return
				}
				fmt.Println("looking for updates at", t)
				asset.AssetVersion = GetVersion(asset.TargetFolder, asset.AssetName)
				newUpdates, updateFound, err := asset.CheckForUpdates()
				if err != nil {
					fmt.Println(err)
					break
				}
				if !updateFound {
					fmt.Println("no new updates for ", asset.AssetName, " available")
					break
				}
				asset.PrintUpdates(newUpdates)
				if canExecuteUpdate, err := executeUpdateCallback(); err != nil {
					fmt.Println(err)
				} else if canExecuteUpdate {
					newVersion, _, err := asset.Update()
					if err != nil {
						fmt.Println(err)
						break
					}
					fmt.Println("updated ", asset.AssetName, "to", newVersion)
					if err = executeAfterUpdateCallback(); err != nil {
						fmt.Println(err)
						break
					}
				} else {
					fmt.Println("Update not executed: executeUpdateCallback returned 'false'")
				}
			}
		}
	}()
	return
}

//PrintUpdates
//Prints information on given updates.
func (asset Asset) PrintUpdates(updates []UpdateInfo) {
	for _, update := range updates {
		fmt.Println("New update for ", asset.AssetName, " ", asset.AssetVersion, " ---> ", update.Version)
		fmt.Println("Type: ", update.Type)
		fmt.Println("Path: ", update.Path)
		fmt.Println()
	}
}
