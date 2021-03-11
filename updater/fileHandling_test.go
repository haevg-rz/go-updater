package updater

import (
	"fmt"
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
	httpImplementation = server.Client()
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

/*getPathToLatestMajor example: MyApp\beta\latest.txt -> pointing to the latest major
func (asset Asset) TestGetPathToLatestMajor(t *testing.T) (majorPath string) {
	return filepath.Join(asset.AssetName, asset.Channel, latestFileName)
}
*/

func TestAsset_getMajorPath(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		major string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantMajorPath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  "",
		}, args{
			major: "2",
		}, filepath.Join("MyApp", "beta", "2")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotMajorPath := asset.getMajorPath(tt.args.major); gotMajorPath != tt.wantMajorPath {
				t.Errorf("getMajorPath() = %v, want %v", gotMajorPath, tt.wantMajorPath)
			}
		})
	}
}

func TestAsset_getPathToAssetBackUpFile(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		assetFilePath string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantAssetBackUpFile string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  "",
		}, args{
			assetFilePath: filepath.Join("installed", "MyApp", "MyApp.exe"),
		}, filepath.Join("installed", "MyApp", "MyApp.exe.old")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotAssetBackUpFile := asset.getPathToAssetBackUpFile(tt.args.assetFilePath); gotAssetBackUpFile != tt.wantAssetBackUpFile {
				t.Errorf("getPathToAssetBackUpFile() = %v, want %v", gotAssetBackUpFile, tt.wantAssetBackUpFile)
			}
		})
	}
}

func TestAsset_getPathToAssetFile(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		fileExt string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantAssetFilePath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  "",
		}, args{
			fileExt: ".exe",
		}, fmt.Sprint("MyApp", ".exe")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotAssetFilePath := asset.getPathToAssetFile(tt.args.fileExt); gotAssetFilePath != tt.wantAssetFilePath {
				t.Errorf("getPathToAssetFile() = %v, want %v", gotAssetFilePath, tt.wantAssetFilePath)
			}
		})
	}
}

func TestAsset_getPathToImportedUpdateFile(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		cdnUpdateFile string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantLocalUpdateFile string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  "",
		}, args{
			cdnUpdateFile: filepath.Join("MyApp", "beta", "2", "MyApp_x64_2.4.2.exe"),
		}, "update_MyApp_x64_2.4.2.exe"},
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, args{
			cdnUpdateFile: filepath.Join("MyApp", "beta", "2", "MyApp_x64_2.4.2.exe"),
		}, filepath.Join("installed", "MyApp", "update_MyApp_x64_2.4.2.exe")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotLocalUpdateFile := asset.getPathToImportedUpdateFile(tt.args.cdnUpdateFile); gotLocalUpdateFile != tt.wantLocalUpdateFile {
				t.Errorf("getPathToImportedUpdateFile() = %v, want %v", gotLocalUpdateFile, tt.wantLocalUpdateFile)
			}
		})
	}
}

func TestAsset_getPathToLatestMajor(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	tests := []struct {
		name            string
		fields          fields
		wantLatestMajor string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, filepath.Join("MyApp", "beta", "latest.txt")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotMajorPath := asset.getPathToLatestMajor(); gotMajorPath != tt.wantLatestMajor {
				t.Errorf("getPathToLatestMajor() = %v, want %v", gotMajorPath, tt.wantLatestMajor)
			}
		})
	}
}

func TestAsset_getPathToLatestPatchInMajorDir(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		major string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantMajorPath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, args{major: "3"}, filepath.Join("MyApp", "beta", "3", "latest.txt")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotMajorPath := asset.getPathToLatestPatchInMajorDir(tt.args.major); gotMajorPath != tt.wantMajorPath {
				t.Errorf("getPathToLatestPatchInMajorDir() = %v, want %v", gotMajorPath, tt.wantMajorPath)
			}
		})
	}
}

func TestAsset_getPathToVersionJson(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		major       string
		latestMinor string
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantVersionJsonPath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, args{
			major:       "3",
			latestMinor: "3.3.4",
		}, filepath.Join("MyApp", "beta", "3", "3.3.4.json")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotVersionJsonPath := asset.getPathToCdnVersionJson(tt.args.major, tt.args.latestMinor); gotVersionJsonPath != tt.wantVersionJsonPath {
				t.Errorf("getPathToCdnVersionJson() = %v, want %v", gotVersionJsonPath, tt.wantVersionJsonPath)
			}
		})
	}
}

func TestAsset_getCdnSigPath(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		cdnUpdateFile string
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantCdnSigPath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, args{
			cdnUpdateFile: filepath.Join("MyApp", "update_MyApp_2.4.2.exe"),
		}, filepath.Join("MyApp", "update_MyApp_2.4.2.exe.minisig")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotCdnSigPath := asset.getCdnSigPath(tt.args.cdnUpdateFile); gotCdnSigPath != tt.wantCdnSigPath {
				t.Errorf("getCdnSigPath() = %v, want %v", gotCdnSigPath, tt.wantCdnSigPath)
			}
		})
	}
}

func TestAsset_getLocalSigPath(t *testing.T) {
	type fields struct {
		AssetName     string
		AssetVersion  string
		Channel       string
		Client        Client
		DoMajorUpdate bool
		Specs         map[string]string
		TargetFolder  string
	}
	type args struct {
		localUpdateFile string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		wantLocalSigPath string
	}{
		{"default successful", fields{
			AssetName:     "MyApp",
			AssetVersion:  "2.0.0",
			Channel:       "beta",
			Client:        nil,
			DoMajorUpdate: true,
			Specs:         nil,
			TargetFolder:  filepath.Join("installed", "MyApp"),
		}, args{
			localUpdateFile: filepath.Join("installed", "MyApp", "update_MyApp_2.4.2.exe"),
		}, filepath.Join("installed", "MyApp", "update_MyApp_2.4.2.exe.minisig")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := Asset{
				AssetName:     tt.fields.AssetName,
				AssetVersion:  tt.fields.AssetVersion,
				Channel:       tt.fields.Channel,
				Client:        tt.fields.Client,
				DoMajorUpdate: tt.fields.DoMajorUpdate,
				Specs:         tt.fields.Specs,
				TargetFolder:  tt.fields.TargetFolder,
			}
			if gotLocalSigPath := asset.getLocalSigPath(tt.args.localUpdateFile); gotLocalSigPath != tt.wantLocalSigPath {
				t.Errorf("getLocalSigPath() = %v, want %v", gotLocalSigPath, tt.wantLocalSigPath)
			}
		})
	}
}

func Test_getPathToLocalVersionJson(t *testing.T) {
	type args struct {
		assetName    string
		targetFolder string
	}
	tests := []struct {
		name                    string
		args                    args
		wantVersionJsonFilePath string
	}{
		{
			name: "default successful",
			args: args{
				assetName:    "MyApp",
				targetFolder: filepath.Join("installed", "MyApp"),
			},
			wantVersionJsonFilePath: filepath.Join("installed", "MyApp", "MyApp_Version.json"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotVersionJsonFilePath := getPathToLocalVersionJson(tt.args.assetName, tt.args.targetFolder); gotVersionJsonFilePath != tt.wantVersionJsonFilePath {
				t.Errorf("getPathToLocalVersionJson() = %v, want %v", gotVersionJsonFilePath, tt.wantVersionJsonFilePath)
			}
		})
	}
}
