# go-updater
Go package for auto-updating binaries and other assets via HTTP Fileserver (Students project)

## Install

`go get github.com/haevg-rz/go-updater/update`

## Feature set

- Self update (of running executable or deamon/services)
- Automatic updating :clock2: 
- Check for update :eyes: 
- Optional no major version update :guardsman: 
- Updating of external assets (with optional compression) :floppy_disk: 
- every asset is signed with Ed25519 :lock: 
- Support of different asset version (like windows, linux) :apple: :lemon: 
- Delegate to check if update is allowed or skipped :question:
- Only a :earth_africa: CDN or :computer: FileShare is needed

**Upload Tool**

- Different targets
  - FileShare
  - Azure CDN
  - more to come
- Signing of assets

## Use Case

```
myapp.exe -> https://example.org/myapp/1/myapp_win_amd64.zip

(.NET 5 App)
- /shippedFirstDotNetApp/mydotnetapp.exe  -> https://example.org/shippedFirstDotNetApp/beta/1/shippedFirstDotNetApp_win_amd64.zip
- /shippedFirstDotNetApp/*.dll

(Single file Publish)
- /shippedSecondDotNetApp/mydotnetapp.exe -> https://example.org/shippedSecondDotNetApp/beta/1/shippedSecondDotNetApp_win_amd64.exe

- /databases/database_customer_xyz.sqlite -> https://example.org/shippedSecondDotNetApp/beta/1/database_customer_xyz.sqlite
```

`myapp.exe` uses this package to update itself and the .NET application `mydotnetapp.exe` and its dependencies and the database `database_customer_xyz.sqlite`.

## File Structure

```
https://example.org/{AssetName}/{Channel}/latest.txt pointing to the latest major, eg. "1"
https://example.org/{AssetName}/{Channel}/{Major}/latest.txt pointing to the latest minor or patch, eg. "1.2.3"
https://example.org/{AssetName}/{Channel}/{Major}/{AssetName}_{Version}_{Specs}_{FileExtension} the actual major´s minor or patch file
```

### Example

**TOOO**

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

	// Do a check for an updater
	_, _ = assetApp.CheckForUpdate()

	// Do a self updater
	_, _ = assetApp.SelfUpdate()

	// Check is a previously updater was aborted
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

	// Do a self updater
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

	// Do a self updater
	_, _ = assetDb.SelfUpdate()
}

func getVersion() string {
	return "0.0.1"
}
```

## Upload tool


