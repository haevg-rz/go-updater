package updater

import (
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type batchData struct {
	ProgramName    string
	DeprecatedName string
	UpdateFileName string
	BatchFileName  string
}

const (
	batchFileName = "updater.bat"
	batchScript   = `Taskkill /IM {{.ProgramName}} /F
	rename {{.ProgramName}} {{.DeprecatedName}}
	rename {{.UpdateFileName}} {{.ProgramName}}
	start {{.ProgramName}}
	del {{.BatchFileName}}
	`
)

func (a Asset) applySelfUpdate(updateFile string) error {
	if err := writeSelfUpdateBatch(updateFile); err != nil {
		return err
	}
	return runWindowsBatch(batchFileName)
}

func writeSelfUpdateBatch(updateFile string) (err error) {
	file, err := os.Create(batchFileName)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			return
		}
	}()
	batchTemplate, err := template.New("batch").Parse(batchScript)
	if err != nil {
		return err
	}
	parameter := batchData{
		ProgramName:    filepath.Base(os.Args[0]),
		DeprecatedName: filepath.Base(os.Args[0]) + ".old",
		UpdateFileName: updateFile,
		BatchFileName:  batchFileName,
	}
	return batchTemplate.Execute(file, parameter)
}

func runWindowsBatch(batchFile string) error {
	cmd := exec.Command("cmd", "/c", batchFile)
	return cmd.Start()
}

func (a Asset) applyUpdate(localUpdateFile string) (err error) {
	fileExt := filepath.Ext(localUpdateFile)
	assetFile := a.getPathToAssetFile(fileExt)
	backUpFile := a.getPathToAssetBackUpFile(assetFile)

	if err = os.Rename(assetFile, backUpFile); err != nil {
		return err
	}
	if err = os.Rename(localUpdateFile, assetFile); err != nil {
		return err
	}
	return nil
}
