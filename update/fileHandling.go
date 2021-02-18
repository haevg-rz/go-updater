package updater

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

const httpIdentifier = "http"

func (asset Asset) read(location string) (content []byte, err error) {

	cdnBaseUrl, err := url.Parse(asset.CdnBaseUrl)
	if err != nil {
		return
	}
	if strings.Contains(cdnBaseUrl.Scheme, httpIdentifier) {
		return asset.readHttp(location)
	}
	return asset.readLocal(location)
}

func (asset Asset) readLocal(location string) (content []byte, err error) {
	location = filepath.Join(asset.CdnBaseUrl, location)
	return ioutil.ReadFile(location)
}

func (asset Asset) readHttp(location string) (content []byte, err error) {
	location = filepath.ToSlash(location)
	u, err := url.Parse(asset.CdnBaseUrl)
	if err != nil {
		return
	}
	u.Path = path.Join(u.Path, location)
	location = u.String()

	client := http.Client{}
	req, err := http.NewRequest("GET", location, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	return ioutil.ReadAll(resp.Body)
}

func (asset Asset) createMajorPath(major string) (majorPath string) {
	return filepath.Join(asset.AssetName, asset.Channel, major)
}
