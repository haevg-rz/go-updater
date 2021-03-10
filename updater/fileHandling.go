package updater

import (
	"encoding/json"
	"fmt"
	"github.com/artdarek/go-unzip"
	"github.com/jedisct1/go-minisign"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
)

type HttpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	HttpImplementation HttpClientInterface
)

func init() {
	HttpImplementation = &http.Client{}
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
	location, err := combineUrlAndFilePathToUrl(HttpClient.CdnBaseUrl, location)
	if err != nil {
		return nil, err
	}
	return readHttpGetRequest(location, HttpImplementation)
}

func combineUrlAndFilePathToUrl(cdnBaseUrl string, location string) (URL string, err error) {
	location = filepath.ToSlash(location)
	u, err := url.Parse(cdnBaseUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, location)
	return u.String(), nil
}

func readHttpGetRequest(location string, client HttpClientInterface) ([]byte, error) {
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

func (asset Asset) importFile(src string, dest string) (err error) {
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

//getLocalSigPath example: installed\MyApp\update_MyApp_2.4.2.exe.minisig
func (asset Asset) getLocalSigPath(localUpdateFile string) (localSigPath string) {
	const signatureSuffix = ".minisig"
	return fmt.Sprint(localUpdateFile, signatureSuffix)
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
	return fmt.Sprint(assetFilePath, backUpSuffix)
}

//getPathToLocalVersionJson example: installed\MyApp\beta\2\MyApp_Version.json
func (asset Asset) getPathToLocalVersionJson() (versionJsonFilePath string) {
	const versionJsonEnding = "_Version.json"
	fileName := fmt.Sprint(asset.AssetName, versionJsonEnding)
	return filepath.Join(asset.TargetFolder, fileName)
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
	versionJsonPath := asset.getPathToLocalVersionJson()
	versionJson := &struct{ Version string }{Version: version}
	content, err := json.Marshal(versionJson)
	if err != nil {
		return
	}
	return ioutil.WriteFile(versionJsonPath, content, 0644)
}

func isSignatureValid(fileName string, signatureFile string) (sigValid bool, err error) {
	const pubKeyFile = "minisign.pub"
	pub, err := minisign.NewPublicKeyFromFile(pubKeyFile)
	if err != nil {
		return
	}
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	sig, err := minisign.NewSignatureFromFile(signatureFile)
	if err != nil {
		return
	}
	return pub.Verify(file, sig)
}
