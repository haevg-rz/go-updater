package updater

import (
	"encoding/json"
	"fmt"
	"github.com/artdarek/go-unzip"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"
)

type httpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	httpImplementation httpClientInterface
)

func init() {
	httpImplementation = &http.Client{Timeout: 10 * time.Second}
}

type Client interface {
	readData(location string) (data []byte, err error)
}

type HttpClient struct {
	CdnBaseUrl string
}

type LocalClient struct {
	CdnBaseUrl string
}

func (HttpClient HttpClient) readData(location string) ([]byte, error) {
	location, err := getTargetUrl(HttpClient.CdnBaseUrl, location)
	if err != nil {
		return nil, err
	}
	return readHttpGetRequest(location, httpImplementation)
}

func getTargetUrl(cdnBaseUrl string, location string) (Url string, err error) {
	location = filepath.ToSlash(location)
	u, err := url.Parse(cdnBaseUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, location)
	return u.String(), nil
}

func readHttpGetRequest(location string, client httpClientInterface) ([]byte, error) {
	req, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func (LocalClient LocalClient) readData(location string) (data []byte, err error) {
	return ioutil.ReadFile(filepath.Join(LocalClient.CdnBaseUrl, location))
}

//TODO Reader / writer pattern (interface)
func (asset Asset) saveRemoteFile(src string, dest string) (err error) {
	data, err := asset.Client.readData(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, data, 0644)
}

//getPathToLatestMajor example: MyApp\beta\latest.txt -> pointing to the latest major
func (asset Asset) getPathToLatestMajor() (latestMajor string) {
	return filepath.Join(asset.AssetName, asset.Channel, latestFileName)
}

//getMajorPath example: MyApp\beta\3\... -> containing updates of this major
func (asset Asset) getMajorPath(major string) (majorPath string) {
	return filepath.Join(asset.AssetName, asset.Channel, major)
}

//getPathToLatestPatchInMajorDir example: MyApp\beta\3\latest.txt -> pointing to the latest patch or minor
func (asset Asset) getPathToLatestPatchInMajorDir(major string) (latest string) {
	return filepath.Join(asset.getMajorPath(major), latestFileName)
}

//getPathToCdnVersionJson example: MyApp\beta\3\3.5.12.json -> containing meta information on 3.5.12 updates
func (asset Asset) getPathToCdnVersionJson(major string, latestMinor string) (versionJsonPath string) {
	const jsonFileExtension = ".json"
	majorPath := asset.getMajorPath(major)
	jsonFileName := fmt.Sprint(latestMinor, jsonFileExtension)
	return filepath.Join(majorPath, jsonFileName)
}

//getPathToImportedUpdateFile example: installed\MyApp\update_MyApp_2.4.2.exe
func (asset Asset) getPathToImportedUpdateFile(cdnUpdateFile string) (localUpdateFile string) {
	const updatePrefix = "update_"
	cdnUpdateFileName := filepath.Base(cdnUpdateFile)
	localUpdateFileName := fmt.Sprint(updatePrefix, cdnUpdateFileName)
	return filepath.Join(asset.TargetFolder, localUpdateFileName)
}

//getCdnSigPath example: MyApp\beta\2\MyApp_2.4.2.exe.minisig
func (asset Asset) getCdnSigPath(cdnUpdateFile string) (cdnSigPath string) {
	const signatureSuffix = ".minisig"
	return fmt.Sprint(cdnUpdateFile, signatureSuffix)
}

//getPathToAssetFile example: installed\MyApp\beta\2\MyApp_2.4.2.exe
func (asset Asset) getPathToAssetFile(fileExt string) (assetFilePath string) {
	assetFile := fmt.Sprint(asset.AssetName, fileExt)
	return filepath.Join(asset.TargetFolder, assetFile)
}

//getPathToAssetBackUpFile example: installed\MyApp\beta\2\MyApp_2.4.2.exe.old
func (asset Asset) getPathToAssetBackUpFile(assetFilePath string) (assetBackUpFile string) {
	const backUpSuffix = ".old"
	return assetFilePath + backUpSuffix
}

//getPathToLocalVersionJson example: installed\MyApp\beta\2\MyApp_Version.json
func getPathToLocalVersionJson(assetName string, targetFolder string) (versionJsonFilePath string) {
	const versionJsonEnding = "_Version.json"
	fileName := fmt.Sprint(assetName, versionJsonEnding)
	return filepath.Join(targetFolder, fileName)
}

func unzipIfCompressed(updatePath string, zipSource string, zipDestination string) (err error) {
	const compressedFileExtension = ".zip"
	if fileExtension := filepath.Ext(updatePath); fileExtension == compressedFileExtension {
		uz := unzip.New(zipSource, zipDestination)
		err = uz.Extract()
	}
	return err
}

func (asset Asset) writeVersionJson(version string) (err error) {
	versionJsonPath := getPathToLocalVersionJson(asset.AssetName, asset.TargetFolder)
	versionJson := &struct{ Version string }{Version: version}
	content, err := json.Marshal(versionJson)
	if err != nil {
		return
	}
	return ioutil.WriteFile(versionJsonPath, content, 0644)
}
