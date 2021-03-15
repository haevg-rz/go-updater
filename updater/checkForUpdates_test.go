package updater

import (
	"path/filepath"
	"testing"
)

func Test_getUpdateType(t *testing.T) {
	type args struct {
		currentVersion string
		newVersion     string
	}
	tests := []struct {
		name           string
		args           args
		wantUpdateType string
		wantErr        bool
	}{
		{
			name: "invalid current version",
			args: args{
				currentVersion: "v1",
				newVersion:     "2.0.0",
			},
			wantUpdateType: "",
			wantErr:        true,
		},
		{
			name: "invalid new version",
			args: args{
				currentVersion: "1.0.1",
				newVersion:     "",
			},
			wantUpdateType: "",
			wantErr:        true,
		},
		{
			name: "major update",
			args: args{
				currentVersion: "1.0.1",
				newVersion:     "2.0.0",
			},
			wantUpdateType: "major",
			wantErr:        false,
		},
		{
			name: "minor update",
			args: args{
				currentVersion: "1.0.1",
				newVersion:     "1.2.0",
			},
			wantUpdateType: "minor",
			wantErr:        false,
		},
		{
			name: "patch update",
			args: args{
				currentVersion: "1.0.1",
				newVersion:     "1.0.2",
			},
			wantUpdateType: "patch",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdateType, err := getUpdateType(tt.args.currentVersion, tt.args.newVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUpdateType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdateType != tt.wantUpdateType {
				t.Errorf("getUpdateType() gotUpdateType = %v, want %v", gotUpdateType, tt.wantUpdateType)
			}
		})
	}
}

func TestAsset_isUpdateValid(t *testing.T) {
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
		availableUpdate AvailableUpdate
		latest          string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMatch bool
	}{
		{
			name: "default successful",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "MyApp",
					Channel: "beta",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "x64",
						"Platform":     "windows",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.1",
			},
			wantMatch: true,
		},
		{
			name: "asset name does not match",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "OtherApp",
					Channel: "beta",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "x64",
						"Platform":     "windows",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.1",
			},
			wantMatch: false,
		},
		{
			name: "channel does not match",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "MyApp",
					Channel: "stable",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "x64",
						"Platform":     "windows",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.1",
			},
			wantMatch: false,
		},
		{
			name: "available update is not latest",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "MyApp",
					Channel: "beta",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "x64",
						"Platform":     "windows",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.2",
			},
			wantMatch: false,
		},
		{
			name: "non equivalent amount of specs",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "MyApp",
					Channel: "beta",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "x64",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.1",
			},
			wantMatch: false,
		},
		{
			name: "specifications do not match",
			fields: fields{
				AssetName:     "MyApp",
				AssetVersion:  "1.0.0",
				Channel:       "beta",
				Client:        nil,
				DoMajorUpdate: true,
				Specs: map[string]string{
					"Architecture": "x64",
					"Platform":     "windows",
				},
				TargetFolder: "",
			},
			args: args{
				availableUpdate: AvailableUpdate{
					Asset:   "MyApp",
					Channel: "beta",
					Version: "1.0.1",
					Specs: map[string]string{
						"Architecture": "arm",
						"Platform":     "linux",
					},
					FilePath: filepath.Join("MyApp", "beta", "1", "MyApp"),
				},
				latest: "1.0.1",
			},
			wantMatch: false,
		},
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
			if gotMatch := asset.isUpdateValid(tt.args.availableUpdate, tt.args.latest); gotMatch != tt.wantMatch {
				t.Errorf("isUpdateValid() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}
