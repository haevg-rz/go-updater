package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"
)

const zipExt = ".zip"

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

type AzureHttpClient struct {
	AccountName string
	AccountKey  string
	CdnBaseUrl  string
	Container   string
	UpdatesRoot string
}

func (a AzureHttpClient) readData(location string) ([]byte, error) {
	location = filepath.Join(a.UpdatesRoot, location)
	location = filepath.ToSlash(location)

	credential, err := azblob.NewSharedKeyCredential(a.AccountName, a.AccountKey)
	if err != nil {
		return nil, err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, err := url.Parse(a.CdnBaseUrl)
	if err != nil {
		return nil, err
	}

	serviceURL := azblob.NewServiceURL(*u, p)
	ctx := context.Background()
	containerURL := serviceURL.NewContainerURL(a.Container)
	blobURL := containerURL.NewBlockBlobURL(location)

	get, err := blobURL.Download(ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	reader := get.Body(azblob.RetryReaderOptions{})
	return ioutil.ReadAll(reader)
}

func (h HttpClient) readData(location string) ([]byte, error) {
	location, err := getTargetUrl(h.CdnBaseUrl, location)
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

func (l LocalClient) readData(location string) (data []byte, err error) {
	return ioutil.ReadFile(filepath.Join(l.CdnBaseUrl, location))
}

//TODO Reader / writer pattern (interface)
func (a Asset) saveRemoteFile(src string, dest string) (err error) {
	data, err := a.Client.readData(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, data, 0644)
}

func (a Asset) writeVersionJson(version string) (err error) {
	versionJsonPath := getPathToLocalVersionJson(a.AssetName, a.TargetFolder)
	versionJson := &struct{ Version string }{Version: version}
	content, err := json.Marshal(versionJson)
	if err != nil {
		return
	}
	return ioutil.WriteFile(versionJsonPath, content, 0644)
}

//getPathToLatestMajor example: MyApp\beta\latest.txt -> pointing to the latest major
func (a Asset) getPathToLatestMajor() (latestMajor string) {
	return filepath.Join(a.AssetName, a.Channel, latestFileName)
}

//getMajorPath example: MyApp\beta\3\... -> containing updates of this major
func (a Asset) getMajorPath(major string) (majorPath string) {
	return filepath.Join(a.AssetName, a.Channel, major)
}

//getPathToLatestPatchInMajorDir example: MyApp\beta\3\latest.txt -> pointing to the latest patch or minor
func (a Asset) getPathToLatestPatchInMajorDir(major string) (latest string) {
	return filepath.Join(a.getMajorPath(major), latestFileName)
}

//getPathToCdnVersionJson example: MyApp\beta\3\3.5.12.json -> containing meta information on 3.5.12 updates
func (a Asset) getPathToCdnVersionJson(major string, latestMinor string) (versionJsonPath string) {
	const jsonFileExtension = ".json"
	majorPath := a.getMajorPath(major)
	jsonFileName := fmt.Sprint(latestMinor, jsonFileExtension)
	return filepath.Join(majorPath, jsonFileName)
}

//getPathToLocalUpdateFile example: installed\MyApp\update_MyApp_2.4.2.exe
func (a Asset) getPathToLocalUpdateFile(cdnUpdateFile string) (localUpdateFile string) {
	const updatePrefix = "update_"
	cdnUpdateFileName := filepath.Base(cdnUpdateFile)
	localUpdateFileName := fmt.Sprint(updatePrefix, cdnUpdateFileName)
	if fileExt := filepath.Ext(cdnUpdateFile); fileExt == zipExt {
		targetFolderParentDir := filepath.Dir(a.TargetFolder)
		return filepath.Join(targetFolderParentDir, localUpdateFileName)
	} else {
		return filepath.Join(a.TargetFolder, localUpdateFileName)
	}
}

//getCdnSigPath example: MyApp\beta\2\MyApp_2.4.2.exe.minisig
func (a Asset) getCdnSigPath(cdnUpdateFile string) (cdnSigPath string) {
	const signatureSuffix = ".minisig"
	return fmt.Sprint(cdnUpdateFile, signatureSuffix)
}

//getPathToAssetFile example: installed\MyApp\beta\2\MyApp_2.4.2.exe
func (a Asset) getPathToAssetFile(fileExt string) (assetFilePath string) {
	if fileExt == zipExt {
		return a.TargetFolder
	} else {
		assetFile := a.AssetName + fileExt
		return filepath.Join(a.TargetFolder, assetFile)
	}
}

//getPathToAssetBackUpFile example: installed\MyApp\beta\2\MyApp_2.4.2.exe.old
func (a Asset) getPathToAssetBackUpFile(assetFilePath string) (assetBackUpFile string) {
	const backUpSuffix = ".old"
	return assetFilePath + backUpSuffix
}

//getPathToLocalVersionJson example: installed\MyApp\beta\2\MyApp_Version.json
func getPathToLocalVersionJson(assetName string, targetFolder string) (versionJsonFilePath string) {
	const versionJsonEnding = "_Version.json"
	fileName := fmt.Sprint(assetName, versionJsonEnding)
	return filepath.Join(targetFolder, fileName)
}
