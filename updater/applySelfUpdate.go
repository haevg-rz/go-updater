package updater

import (
	"fmt"
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
	batchFileName  = "updater.bat"
	updateFileName = "updater"
	batchScript    = `Taskkill /IM {{.ProgramName}} /F
	rename {{.ProgramName}} {{.DeprecatedName}}
	rename {{.UpdateFileName}} {{.ProgramName}}
	start {{.ProgramName}}
	del {{.BatchFileName}}
	`
)

func (asset Asset) applySelfUpdate() error {
	if err := writeBatchFile(); err != nil {
		return err
	}
	return runWindowsBatch(batchFileName)
}

func writeBatchFile() (err error) {
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
		DeprecatedName: fmt.Sprint(filepath.Base(os.Args[0]), ".old"),
		UpdateFileName: updateFileName,
		BatchFileName:  batchFileName,
	}
	return batchTemplate.Execute(file, parameter)
}

func runWindowsBatch(batchFile string) error {
	cmd := exec.Command("cmd", "/c", batchFile)
	return cmd.Start()
}
