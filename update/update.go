//Welcome to the updater Module

//This file contains the exposed Functions. If you are new to the project and want to understand the code,
//its recommended to start here!

package updater

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type Asset struct {
	AssetName     string
	AssetVersion  string
	Channel       string
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
// Looks for the latest available updates. Applies the newest update, terminating the running process and exchanging the executable files. Then restarts the application.
func (asset Asset) SelfUpdate() (updatedTo UpdateInfo, err error) {
	availableUpdates, err := asset.CheckForUpdates()
	if err != nil {
		return
	}
	latestUpdate := asset.getLatestAllowedUpdate(availableUpdates)
	err = asset.importSelfUpdate(latestUpdate.Path)
	if err != nil {
		return
	}
	err = asset.applySelfUpdate()
	if err != nil {
		return
	}
	updatedTo = latestUpdate
	return
}

// Update
// Looks for the latest available updates of an external Asset. Applies the newest update and writes a versionJson into the asset folder, which points to the new version.
func (asset Asset) Update() (updatedTo UpdateInfo, err error) {
	availableUpdates, err := asset.CheckForUpdates()
	if err != nil {
		return
	}
	latestUpdate := asset.getLatestAllowedUpdate(availableUpdates)
	err = asset.importUpdate(latestUpdate.Path)
	if err != nil {
		return
	}
	updatedTo = latestUpdate
	asset.AssetVersion = updatedTo.Version
	err = asset.writeVersionJson(latestUpdate.Version)
	return
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

//TODO silent Mode for printing
//Background
//Starts looking for updates in a specified interval. Use the allowUpdate function to enable/disable updates.
func (asset Asset) Background(interval time.Duration, allowUpdate func() bool) (err error) {
	ticker := time.NewTicker(interval)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				asset.AssetVersion = GetVersion(asset.TargetFolder, asset.AssetName)
				newUpdates, err := asset.CheckForUpdates()
				printErrors(err)
				asset.PrintUpdates(newUpdates)
				if allowUpdate() {
					_, err = asset.Update()
					printErrors(err)
				}
			}
		}
	}()
	time.Sleep(time.Second * 60)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
	return
}

//PrintUpdates
//Prints information on given updates.
func (asset Asset) PrintUpdates(updates []UpdateInfo) {
	for _, update := range updates {
		println(fmt.Sprint(fmt.Sprint("New update for ", asset.AssetName, " ", asset.AssetVersion, " ---> ", update.Version)))
		println(fmt.Sprint("Type: ", update.Type))
		println(fmt.Sprint("Path: ", update.Path), "\n")
	}
}
