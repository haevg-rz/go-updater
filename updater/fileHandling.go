package updater

import (
	"fmt"
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

//latestMajorPath example: MyApp\beta\latest.txt -> pointing to the latest major
func (asset Asset) getPathToLatestMajor() (majorPath string) {
	return filepath.Join(asset.AssetName, asset.Channel, latestFileName)
}

//getMajorPath example: MyApp\beta\3\... -> containing updates of this major
func (asset Asset) getMajorPath(major string) (majorPath string) {
	return filepath.Join(asset.AssetName, asset.Channel, major)
}

//getPathToLatestPatchInMajorDir example: MyApp\beta\3\latest.txt -> pointing to the latest patch or minor
func (asset Asset) getPathToLatestPatchInMajorDir(major string) (majorPath string) {
	return filepath.Join(asset.getMajorPath(major), latestFileName)
}

//getPathToVersionJson example: MyApp\beta\3\3.5.12.json -> containing meta information on 3.5.12 updates
func (asset Asset) getPathToVersionJson(major string, latestMinor string) (versionJsonPath string) {
	const jsonFileExtension = ".json"
	majorPath := asset.getMajorPath(major)
	jsonFileName := fmt.Sprint(latestMinor, jsonFileExtension)
	return filepath.Join(majorPath, jsonFileName)
}
