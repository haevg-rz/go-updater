package updater

import (
	"github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalClientReadData(t *testing.T) {
	//arrange
	defer filet.CleanUp(t)
	CdnBaseUrl, _ := os.Getwd()
	testFile := "go-testLocalClientRead.txt"
	expected := []byte("Successful")
	filet.File(t, filepath.Join(CdnBaseUrl, testFile), string(expected))
	var testAsset = Asset{
		Client: LocalClient{CdnBaseUrl: CdnBaseUrl},
	}
	//act
	got, err := testAsset.Client.readData(testFile)
	if err != nil {
		log.Fatal(err)
	}
	//assert
	assert.Equal(t, string(expected), string(got))
}

func TestHttpClientReadData(t *testing.T) {
	//arrange
	expected := "1.0.0"
	testFileLocation := filepath.Join("/MyApp", "Beta", "1", "latest.txt")
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.String() == filepath.ToSlash(testFileLocation) {
			_, err := rw.Write([]byte(expected))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := rw.Write([]byte("file not found"))
			if err != nil {
				log.Fatal(err)
			}
		}
	}))
	defer server.Close()
	HttpImplementation = server.Client()
	var testAsset = Asset{
		Client: HttpClient{CdnBaseUrl: server.URL},
	}
	//act
	got, _ := testAsset.Client.readData(testFileLocation)
	//assert
	assert.Equal(t, expected, string(got))
}

func TestCombineUrlAndFilePathToUrl(t *testing.T) {
	//arrange
	expected := "https://myStorage.blob.core.windows.net/updatescontainer/Updates/MyApp/Beta/1/latest.txt"
	cdnBaseUrl := "https://myStorage.blob.core.windows.net/updatescontainer/Updates"
	location := filepath.Join("/MyApp", "Beta", "1", "latest.txt")
	//act
	got, _ := combineUrlAndFilePathToUrl(cdnBaseUrl, location)
	//assert
	assert.Equal(t, got, expected)
}

func TestCreateMajorPath(t *testing.T) {
	//arrange
	testAsset := Asset{
		AssetName: "MyApp",
		Channel:   "beta",
	}
	major := "1"
	expected := filepath.Join(testAsset.AssetName, testAsset.Channel, major)
	//act
	got := testAsset.getMajorPath(major)
	//assert
	assert.Equal(t, expected, got)
}
