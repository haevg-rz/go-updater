# go-updater
Go package for auto-updating binaries and other assets via HTTP Fileserver (Students project)

## Install

`go get github.com/haevg-rz/go-updater/update`

## Feature set



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
		CdnBaseURL:    "https://cdn.company.com/updates/",
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

	// Check is a previously update was aborted
	_ = assetApp.UpdateAborted()

	// Start a background goroutine for continuous checks with random
	go assetApp.Background(time.Hour, time.Minute*10, allowUpdate)

	updateDatabaseAsset()
	updateDotNetApp()
}

func allowUpdate() bool {
	return true
}

```

### Update external assets

```go

func updateDatabaseAsset() {
	assetDb := &update.Asset{
		AssetVersion: getVersion(),
		AssetName:    "MyDatabases",
		Channel:      "Stable",
		CdnBaseURL:   "https://cdn.company.com/updates/",
		Specs: map[string]string{
			"Name": "MyContacts",
			"Type": "SQlite",
		},
		TargetFolder: "db",
	}

	// Do a self update
	_, _ = assetDb.SelfUpdate()
}

func getVersion() string {
	return "0.0.1"
}

```

### Update external dot net apps

```go
func updateDotNetApp() {
	assetDb := &update.Asset{
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

	// Do a self update
	_, _ = assetDb.SelfUpdate()
}

func getVersion() string {
	return "0.0.1"
}
```

## Upload tool


