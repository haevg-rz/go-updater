package updater

import "testing"

func Test_getSemanticVersioningParts(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name      string
		args      args
		wantMajor string
		wantMinor string
		wantPatch string
		wantErr   bool
	}{
		{"default successful", args{version: "1.2.3"}, "1", "2", "3", false}, {"invalid version", args{version: "v0"}, "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, gotPatch, err := getSemanticVersioningParts(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSemanticVersioningParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMajor != tt.wantMajor {
				t.Errorf("getSemanticVersioningParts() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotMinor != tt.wantMinor {
				t.Errorf("getSemanticVersioningParts() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
			}
			if gotPatch != tt.wantPatch {
				t.Errorf("getSemanticVersioningParts() gotPatch = %v, want %v", gotPatch, tt.wantPatch)
			}
		})
	}
}

func Test_isUpdateNewerThanCurrent(t *testing.T) {
	type args struct {
		currentVersion string
		updateVersion  string
	}
	tests := []struct {
		name              string
		args              args
		wantUpdateIsNewer bool
		wantErr           bool
	}{
		{
			name: "invalid current version",
			args: args{
				currentVersion: "1.2",
				updateVersion:  "1.0.0",
			},
			wantUpdateIsNewer: false,
			wantErr:           true,
		},
		{
			name: "invalid update version",
			args: args{
				currentVersion: "1.2.0",
				updateVersion:  "1.0",
			},
			wantUpdateIsNewer: false,
			wantErr:           true,
		},
		{
			name: "update major > current major",
			args: args{
				currentVersion: "1.2.0",
				updateVersion:  "2.0.0",
			},
			wantUpdateIsNewer: true,
			wantErr:           false,
		},
		{
			name: "update minor > current minor",
			args: args{
				currentVersion: "1.2.0",
				updateVersion:  "1.3.0",
			},
			wantUpdateIsNewer: true,
			wantErr:           false,
		},
		{
			name: "update patch > current patch",
			args: args{
				currentVersion: "1.2.0",
				updateVersion:  "1.2.1",
			},
			wantUpdateIsNewer: true,
			wantErr:           false,
		},
		{
			name: "update is not newer",
			args: args{
				currentVersion: "1.2.0",
				updateVersion:  "1.2.0",
			},
			wantUpdateIsNewer: false,
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUpdateIsNewer, err := isUpdateNewerThanCurrent(tt.args.currentVersion, tt.args.updateVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("isUpdateNewerThanCurrent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdateIsNewer != tt.wantUpdateIsNewer {
				t.Errorf("isUpdateNewerThanCurrent() gotUpdateIsNewer = %v, want %v", gotUpdateIsNewer, tt.wantUpdateIsNewer)
			}
		})
	}
}
