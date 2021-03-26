# go-updater
Go package for auto-updating binaries and other assets via HTTP Fileserver (Students project)

[![Go](https://github.com/haevg-rz/go-updater/actions/workflows/go.yml/badge.svg)](https://github.com/haevg-rz/go-updater/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/haevg-rz/go-updater/branch/main/graph/badge.svg?token=JFFS77RP56)](https://codecov.io/gh/haevg-rz/go-updater)

## Install
`go get github.com/haevg-rz/go-updater/update`

## Feature set

- Self update (of running executable or deamon/services) 
- Check for update :eyes: 
- Optional no major version update :guardsman: 
- Updating of external assets (with optional compression) :floppy_disk: 
- Support of different asset version (like windows, linux) :apple: :lemon: 
- Only a :earth_africa: CDN or :computer: FileShare is needed
- Delegate to check if update is allowed or skipped :question:
- Automatic updating :clock2:
- every asset is signed with Ed25519 :lock: 

**Upload Tool** (to be implemented)

- Different targets
  - FileShare
  - Azure CDN
  - more to come
- Signing of assets

## Use Case

```
an asset {single file, service, database, entire folder, ...} needs to be updated


## File Structure

```
CDN {http Server, local filesystem}

https://example.org/{AssetName}/{Channel}/latest.txt pointing to the latest major, eg. "1"
https://example.org/{AssetName}/{Channel}/{Major}/latest.txt pointing to the latest minor or patch, eg. "1.2.3"
https://example.org/{AssetName}/{Channel}/{Major}/{version}.json containing metainfo e.g. filepath, assetName, channel to the actual update of this version
https://example.org/{AssetName}/{Channel}/{Major}/{AssetName}_{Version}_{Specs}_{FileExtension} the actual major´s minor or patch update file
https://example.org/{AssetName}/{Channel}/{Major}/{AssetName}_{Version}_{Specs}_{FileExtension}.minisign the signature of the actual update file
```

### Example

see cmd/sample to download a working sample

## Client

```go
package main

import (
	"runtime"
	"time"
	"updatersample/update"
)

var Version = "0.0.0" // Set in build
const AppName = "MyAppName"

func main() {
	var assetApp = &update.Asset{
		AssetVersion:  Version,
		AssetName:     AppName,
		Channel:       "Stable",
		Client:        updater.HttpClient{CdnBaseUrl: "https://example.org"},
		DoMajorUpdate: false,
		Specs: map[string]string{
			"Arch": runtime.GOARCH,
			"OS":   runtime.GOOS,
		},
	}

	// Do a check for an update
	_, _ = assetApp.CheckForUpdate()

	// Do a self update
	_, _ = assetApp.SelfUpdate()

	// Start a background goroutine for continuous checks
	go assetApp.Background(time.Hours * 24, skipAssetAppUpdate, executeAssetAppUpdateCallBack, executeAssetAppAfterUpdateCallBack)

	updateDatabaseAsset()
	updateDotNetApp()
}

```

### Update external assets

```go

func updateDatabaseAsset() {
	assetDb := &update.Asset{
		AssetVersion: updater.getVersion(),
		AssetName:    "MyDatabases",
		Channel:      "Stable",
		Client:       updater.HttpClient{CdnBaseUrl: "https://example.org"},
		Specs: map[string]string{
			"Name": "MyContacts",
			"Type": "SQlite",
		},
		TargetFolder: "db",
	}

	// Update asset
	_, _ = assetDb.Update()
}

```

### Update external dot net apps

```go
func updateDotNetApp() {
	assetDotNet := &update.Asset{
		AssetVersion: getVersion(),
		AssetName:    "MyDotNetApp",
		Channel:      "Stable",
		CdnBaseURL:   "https://cdn.company.com/updates/",
		Specs: map[string]string{
			"Arch":         runtime.GOARCH,
			"OS":           runtime.GOOS,
			"Distribution": "RedHat",
		},
		TargetFolder: "MyDotNetApp",
	}

	// Update asset
	_, _ = assetDotNet.Update()
}

func getVersion() string {
	return "0.0.1"
}
```
