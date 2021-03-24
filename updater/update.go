//Welcome to the updater Module

//This file contains the exposed Functions. If you are new to the project and want to understand the code,
//its recommended to start here!

package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

type Asset struct {
	AssetName     string
	AssetVersion  string
	Channel       string
	Client        Client
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
func (a Asset) SelfUpdate() (updatedTo *UpdateInfo, updated bool, err error) {
	availableUpdates, updateFound, err := a.CheckForUpdates()
	if err != nil {
		return nil, false, err
	}
	if !updateFound {
		return nil, false, nil
	}

	latestUpdate, err := a.getLatestAllowedUpdate(availableUpdates)
	if err != nil {
		return nil, false, err
	}

	localUpdateFile := a.getPathToLocalUpdateFile(latestUpdate.Path)
	cdnSigFile := a.getCdnSigPath(latestUpdate.Path)

	if err = a.saveRemoteFile(latestUpdate.Path, localUpdateFile); err != nil {
		return nil, false, err
	}

	sigValid, err := a.isSignatureValid(localUpdateFile, cdnSigFile)
	if !sigValid || (err != nil) {
		return nil, false, err
	}

	if err = a.applySelfUpdate(localUpdateFile); err != nil {
		return nil, false, err
	}

	return latestUpdate, true, nil
}

// Update
// Looks for the latest available updates of an external Asset. Applies the newest updater and writes a versionJson into the asset folder, which points to the new version.
func (a Asset) Update() (updatedTo *UpdateInfo, updated bool, err error) {
	availableUpdates, updateFound, err := a.CheckForUpdates()
	if err != nil {
		return nil, false, err
	}
	if !updateFound {
		return nil, false, nil
	}

	latestUpdate, err := a.getLatestAllowedUpdate(availableUpdates)
	if err != nil {
		return nil, false, err
	}

	localUpdateFile := a.getPathToLocalUpdateFile(latestUpdate.Path)
	cdnSigFile := a.getCdnSigPath(latestUpdate.Path)

	if err = a.saveRemoteFile(latestUpdate.Path, localUpdateFile); err != nil {
		return nil, false, err
	}

	sigValid, err := a.isSignatureValid(localUpdateFile, cdnSigFile)
	if !sigValid || (err != nil) {
		return nil, false, err
	}

	if err = a.applyUpdate(localUpdateFile); err != nil {
		return nil, false, err
	}

	if err = a.writeVersionJson(latestUpdate.Version); err != nil {
		return nil, false, err
	}
	return latestUpdate, true, nil
}

func (a Asset) getLatestAllowedUpdate(availableUpdates []UpdateInfo) (updateInfo *UpdateInfo, err error) {
	if a.DoMajorUpdate {
		for _, update := range availableUpdates {
			if update.Type == "major" {
				return &update, nil
			}
		}
	}
	for _, update := range availableUpdates {
		if update.Type == "minor" {
			return &update, nil
		}
	}
	for _, update := range availableUpdates {
		if update.Type == "patch" {
			return &update, nil
		}
	}
	return nil, errors.New("no update has an update Type")
}

//GetVersion
//Gets a semantic Versioning string for the asset that is to be updated. Looks for a Version Json, written every time the asset
//is updated by this module. If the json can not be found, a default version of 0.0.0 is returned.
func GetVersion(targetFolder string, assetName string) (currentVersion string) {
	const defaultVersion = "0.0.0"
	currentVersion = defaultVersion
	versionJson := getPathToLocalVersionJson(assetName, targetFolder)
	data, err := ioutil.ReadFile(versionJson)
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
func (a Asset) Background(interval time.Duration, skipUpdate func() bool, executeUpdateCallback func() (bool, error), executeAfterUpdateCallback func() error) (err error) {
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
				a.AssetVersion = GetVersion(a.TargetFolder, a.AssetName)
				newUpdates, updateFound, err := a.CheckForUpdates()
				if err != nil {
					fmt.Println(err)
					break
				}
				if !updateFound {
					fmt.Println("no new updates for ", a.AssetName, " available")
					break
				}
				a.PrintUpdates(newUpdates)
				canExecuteUpdate, err := executeUpdateCallback()
				if err != nil {
					log.Println(err)
				}
				if !canExecuteUpdate {
					fmt.Println("Update not executed: executeUpdateCallback returned 'false'")
					break
				}
				newVersion, _, err := a.Update()
				if err != nil {
					fmt.Println(err)
					break
				}
				fmt.Println("updated ", a.AssetName, "to", newVersion)
				if err = executeAfterUpdateCallback(); err != nil {
					fmt.Println(err)
					break
				}
			}
		}
	}()
	return
}

//PrintUpdates
//Prints information on given updates.
func (a Asset) PrintUpdates(updates []UpdateInfo) {
	for _, update := range updates {
		fmt.Println("New update for ", a.AssetName, " ", a.AssetVersion, " ---> ", update.Version)
		fmt.Println("Type: ", update.Type)
		fmt.Println("Path: ", update.Path)
		fmt.Println()
	}
}
