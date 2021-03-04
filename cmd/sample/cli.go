package main

import (
	"bufio"
	"fmt"
	"github.com/haevg-rz/go-updater/updater"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*Usage
<- Build ->
	=> build the project with ldflags
	-ldflags "-X main.AppName=myCore -X main.Channel=Beta -X main.Platform=windows -X main.Architecture=amd64 -X main.Version=1.0.0"

<- Setup ->
	=> provide updates in an updates directory
		as specified on https://github.com/haevg-rz/go-updater

	=> set the variable CdnBaseUrl to the path of the updates directory

	=> set the TargetFolder variable of assets to update to
		the current directory containing the asset
*/

var (
	AppName      = "unknown" //set on build time
	Channel      = "unknown" //set on build time
	Platform     = "unknown" //set on build time
	Architecture = "unknown" //set on build time
	Version      = "DEV"     //set on build time
	BuildTime    = "unknown" //set on build time

	//CdnBaseUrl
	//Set the root path to the updates directory (manually, programmatically, or on build time)
	CdnBaseUrl = `H:\Entwicklung\Demo Go Updater\UpdatePackage`

	//the concrete
	client = updater.LocalClient{
		CdnBaseUrl: CdnBaseUrl,
	}

	selfUpdateAsset = updater.Asset{
		AssetName:     AppName,
		AssetVersion:  Version,
		Channel:       Channel,
		Client:        client,
		DoMajorUpdate: true,
		Specs: map[string]string{
			"Architecture": runtime.GOARCH,
			"Platform":     runtime.GOOS,
		},
	}

	service1 = updater.Asset{
		AssetName:     "Service1",
		AssetVersion:  updater.GetVersion("H:\\Entwicklung\\Demo Go Updater\\Installed\\Service1", "Service1"),
		Channel:       "Beta",
		Client:        client,
		DoMajorUpdate: true,
		Specs: map[string]string{
			"Architecture": runtime.GOARCH,
			"Platform":     runtime.GOOS,
		},
		TargetFolder: "H:\\Entwicklung\\Demo Go Updater\\Installed\\Service1",
	}

	customerDatabase = updater.Asset{
		AssetName:     "customer_database",
		AssetVersion:  updater.GetVersion("H:\\Entwicklung\\Demo Go Updater\\Installed\\Databases", "customer_database"),
		Channel:       "Beta",
		Client:        client,
		DoMajorUpdate: true,
		TargetFolder:  "H:\\Entwicklung\\Demo Go Updater\\Installed\\Databases",
	}

	reader *bufio.Reader
)

func main() {
	printProgramMetaInfo()
	printStartingMessage()
	startReader()
	readConsoleCommands()
}

func printProgramMetaInfo() {
	fmt.Println("Application:\t", AppName)
	fmt.Println("Channel:\t", Channel)
	fmt.Println("Platform:\t", Platform)
	fmt.Println("Architecture:\t", Architecture)
	fmt.Println("Version:\t", Version)
	fmt.Println("BuildTime:\t", BuildTime)
	fmt.Println("CdnBaseUrl:\t", CdnBaseUrl)
}

func printStartingMessage() {
	const startingMessage = "\nGreetings. This is the myCore app. Update itself or other assets. Type --help to show all commands"
	fmt.Println(startingMessage)
	fmt.Println()
}

func startReader() {
	reader = bufio.NewReader(os.Stdin)
}

func readConsoleCommands() {
	for {
		command, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println("could not process command: ", string(command))
		}
		switch string(command) {
		case "--help":
			fmt.Println("Listing all commands:\n--exit\n--version\n--check self\n--check customer_database\n --check Service1\n --background Service1")
		case "--version":
			printProgramMetaInfo()
		case "--check Service1", "-c Service1":
			CheckForUpdates(service1)
		case "--background Service1", "-b Service1":
			if err := service1.Background(time.Second*10, skipService1BackgroundUpdate, onService1ExecuteUpdate, afterService1BackgroundUpdateExecuted); err != nil {
				fmt.Println(err)
			}
		case "--check customer_database":
			CheckForUpdates(customerDatabase)
		case "--check self":
			CheckForSelfUpdates(selfUpdateAsset)
		case "--exit":
			fmt.Println("exiting program ...")
			os.Exit(0)
		default:
			fmt.Println("unrecognized command: ", string(command))
		}
	}
}

func CheckForUpdates(asset updater.Asset) {
	fmt.Println("checking ", asset.AssetName, " for updates...")
	availableUpdates, updateFound, err := asset.CheckForUpdates()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !updateFound {
		fmt.Println("no updates found")
		return
	}
	asset.PrintUpdates(availableUpdates)
	fmt.Println("Apply Update? - if this asset is a running process, it will be shut down. (y|n) ?")
	input, _, _ := reader.ReadLine()
	if string(input) != "y" {
		fmt.Println("aborted")
		return
	}
	fmt.Println("updating...")
	if isExecutable(availableUpdates[0].Path) {
		if err = killProcess(asset); err != nil {
			fmt.Println(err)
		}
	}
	updatedTo, updated, err := asset.Update()
	if isExecutable(availableUpdates[0].Path) {
		if err := startProcess(asset); err != nil {
			fmt.Println(err)
			return
		}
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	if !updated {
		fmt.Println("could not update", asset.AssetName)
		return
	}
	fmt.Println("successfully updated ", asset.AssetName, " to ", updatedTo)
	return
}

func CheckForSelfUpdates(asset updater.Asset) {
	fmt.Println("checking ", asset.AssetName, " for updates...")
	availableUpdates, updateFound, err := asset.CheckForUpdates()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !updateFound {
		fmt.Println("no updates found")
		return
	}
	asset.PrintUpdates(availableUpdates)
	fmt.Println("apply Update? (y|n) ?")
	input, _, _ := reader.ReadLine()
	if string(input) != "y" {
		fmt.Println("aborted")
		return
	}
	fmt.Println("updating...")
	updatedTo, updated, err := asset.SelfUpdate()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !updated {
		fmt.Println("could not update", asset.AssetName)
		return
	}
	fmt.Println("successfully updated ", asset.AssetName, " to ", updatedTo.Version)
	return
}

func startProcess(asset updater.Asset) error {
	command := fmt.Sprint("cmd /c Start ", asset.AssetName, ".exe")
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = asset.TargetFolder
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func killProcess(asset updater.Asset) error {
	command := fmt.Sprint("cmd /c taskkill /IM ", asset.AssetName, ".exe -F")
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func skipService1BackgroundUpdate() bool {
	return false
}

func onService1ExecuteUpdate() (bool, error) {
	if err := killProcess(service1); err != nil {
		return false, err
	}
	return true, nil
}

func afterService1BackgroundUpdateExecuted() error {
	if err := startProcess(service1); err != nil {
		return err
	}
	return nil
}

func isExecutable(file string) bool {
	return filepath.Ext(file) == ".exe"
}
