package main

import (
	"bufio"
	"fmt"
	"github.com/haevg-rz/go-updater/updater"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/*

<- Information about this file ->
	this file (cli.go) represents a demo application to showcase the features of the go-updater project.
	It is not part of the actual go-updater library.

<- Getting started ->
	>> Updating the sample HelloWorld.txt file <<

		The sample HelloWorld.txt file can be found at
		{workingDirectory}/cmd/sample/installed/HelloWorld/HelloWorld.txt

		Steps
		#1: build this file (cli.go)

			go build cli.go -ldflags "-X github.com/haevg-rz/go-updater/updater.UpdateFilesPubKey=
			RWQWiK4rVAaJQhk42Y8obU0llCEwBcVWwzliy8T0bYj0WC+JZR8xdntY" -o {go-updater directory}

			use -ldflags to set the public Key matching to the private Key which was used to
			encrypt the sample updates

			use -o to build the program into the go-updater folder (root of the project)
		#2: run it. A console should open
		#3: type '--check HelloWorld' into the console
		#4: the console should output the content of HelloWorld.txt which is 'Hello World'
		#5: the user gets an available update from version 1.0.0 -> 1.0.1 displayed
		#6: type 'y' into the console to confirm to apply this update
		#7: the file will be updated
		#8: the console should output the new content of HelloWorld.txt
			which is 'Hello World And Hello Gophers!'
		#9: done!
		Results -->
		#1: {workingDirectory}/cmd/sample/installed/HelloWorld/HelloWorld.txt content changed
			from 'HelloWorld' to 'HelloWorld And Hello Gophers!'
		#2: {workingDirectory}/cmd/sample/installed/HelloWorld/HelloWorld_version.json content
			changed from '1.0.0' to '1.0.1'


<- Build ->
	=> in order to apply any updates, set the publicKey to the matching private key with which
		the signatures were created.
	-ldflags "-X github.com/haevg-rz/go-updater/updater.UpdateFilesPubKey=
			RWQWiK4rVAaJQhk42Y8obU0llCEwBcVWwzliy8T0bYj0WC+JZR8xdntY"

	=> in order to apply self updates, build the project with ldflags
	-ldflags "-X main.AppName=myCore -X main.Channel=Beta -X main.Platform=windows -X main.Architecture=amd64 -X main.Version=1.0.0"


<- Setup your own assets->
	#1: provide updates in an updates directory or on a file server
		as specified on https://github.com/haevg-rz/go-updater and shown in this sample

	#2: create an updater.asset and fill all fields

	#3: set the variable CdnBaseUrl of this assets client to the path of the updates directory.
		for this sample its {workingDirectory}/cmd/sample/updates

	#4: set the TargetFolder variable of this asset, to specify where updates should be applied to.
		The current directory containing the asset HelloWorld is {workingDirectory}/cmd/sample/installed/HelloWorld/HelloWorld.txt
		This is not necessary for self updating.

	#5: run asset.Update() (check and Update), asset.CheckForUpdates() (check only),
	asset.SelfUpdate() (update the program itself), or asset.Background() (update automatically)
	on the asset, depending on the needs of your project
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
	CdnBaseUrl = "https://kaisupdates.blob.core.windows.net/updatescontainer/updates"

	//client
	//Set the type of your cdn, changing the behaviour files are read from the cdn. Http or local file reading.
	client updater.Client

	//asset
	//Create Assets that should be updated
	selfUpdateAsset updater.Asset
	service1        updater.Asset //not usable in this demo
	helloWorld      updater.Asset
	images          updater.Asset

	reader *bufio.Reader
)

func main() {
	printProgramMetaInfo()
	printStartingMessage()

	setUpSamples()

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
	fmt.Println("Updates PubKey:\t", updater.UpdateFilesPubKey)
}

func printStartingMessage() {
	const startingMessage = "\nGreetings. This is the myCore app. Update itself or other assets. Type --help to show all commands"
	fmt.Println(startingMessage)
	fmt.Println()
}

func setUpSamples() {
	wd, _ := os.Getwd()
	//CdnBaseUrl = filepath.Join(wd, "cmd", "sample", "updates")

	client = updater.HttpClient{
		CdnBaseUrl: CdnBaseUrl,
	}

	helloWorld = updater.Asset{
		AssetName:     "HelloWorld",
		AssetVersion:  updater.GetVersion(filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld"), "HelloWorld"),
		Channel:       "Beta",
		Client:        client,
		DoMajorUpdate: true,
		TargetFolder:  filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld"),
	}

	images = updater.Asset{
		AssetName:     "Images",
		AssetVersion:  updater.GetVersion(filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld", "Images"), "Images"),
		Channel:       "Beta",
		Client:        client,
		DoMajorUpdate: true,
		TargetFolder:  filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld", "Images"),
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
}

func startReader() {
	reader = bufio.NewReader(os.Stdin)
}

func readConsoleCommands() {
	commands := []string{"--exit", "--help", "--version", "--check HelloWorld", "--check self", "--bg HelloWorld", "--check Images"}
	for {
		command, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println("could not process command: ", string(command))
		}
		switch string(command) {
		case "--help":
			fmt.Println("Listing all commands:")
			for _, c := range commands {
				fmt.Println(c)
			}
		case "--version":
			printProgramMetaInfo()
		case "--check HelloWorld":
			printHW()
			CheckForUpdates(helloWorld)
			printHW()
		case "--bg HelloWorld":
			helloWorld.Background(time.Second*4, skipHelloWorldUpdate, executeHelloWorldUpdateCallBack, executeHelloWorldAfterUpdateCallBack)
		case "--check Images":
			CheckForUpdates(images)
		case "--check self":
			CheckForSelfUpdates(selfUpdateAsset)
		case "--exit":
			fmt.Println("exiting program ...")
			os.Exit(0)
		/*
			If there is an executable to update automatically, the usage of the Background() function could look like this
			case "--background Service1", "-b Service1":
			if err := service1.Background(time.Second*10, skipService1BackgroundUpdate, onService1ExecuteUpdate, afterService1BackgroundUpdateExecuted); err != nil {
				fmt.Println(err)
			}
		*/
		default:
			fmt.Println("unrecognized command: ", string(command))
		}
	}
}

func printHW() {
	wd, _ := os.Getwd()
	file := filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld", "HelloWorld.txt")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
	}
	text := string(data)
	fmt.Println("file: ", filepath.Join(wd, "cmd", "sample", "installed", "HelloWorld", "HelloWorld.txt"))
	fmt.Println("reads: ", text)
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
	if isExecutable(availableUpdates[0].Path) {
		fmt.Println("Apply Update? - if this asset is a running process, it will be shut down. (y|n) ?")
		input, _, _ := reader.ReadLine()
		if string(input) != "y" {
			fmt.Println("aborted")
			return
		}
		if err = killProcess(asset); err != nil {
			log.Println(err)
		}
	} else {
		fmt.Println("Apply Update? (y|n) ?")
		input, _, _ := reader.ReadLine()
		if string(input) != "y" {
			fmt.Println("aborted")
			return
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
	fmt.Println("successfully updated ", asset.AssetName, " to ", (*updatedTo).Version)
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
	fmt.Println("successfully updated ", asset.AssetName, " to ", (*updatedTo).Version)
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

func isExecutable(file string) bool {
	return filepath.Ext(file) == ".exe"
}

func skipHelloWorldUpdate() bool {
	//logic to skip an update
	return false
}

func executeHelloWorldUpdateCallBack() (bool, error) {
	//actions that should be taken before updating, e.g. stopping a process or db connection
	return true, nil
}

func executeHelloWorldAfterUpdateCallBack() error {
	//actions that should be taken after updating, e.g. starting a process or db connection
	return nil
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
