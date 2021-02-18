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
	batchFileName  = "updater.bat"
	updateFileName = "updater"
	batchScript    = `Taskkill /IM {{.ProgramName}} /F
	rename {{.ProgramName}} {{.DeprecatedName}}
	rename {{.UpdateFileName}} {{.ProgramName}}
	start {{.ProgramName}}
	del {{.BatchFileName}}
	`
)

//TODO handle minisig files
//TODO rename old exe to.old instead of OLD{fileName}.exe
func (asset Asset) applySelfUpdate() error {
	err := writeBatchFile()
	if err != nil {
		return err
	}
	return runWindowsBatch(batchFileName)
}

func writeBatchFile() (err error) {
	err = os.Remove(batchFileName)
	printErrors(err)
	file, err := os.Create(batchFileName)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			return
		}
	}()
	batchTemplate, err := template.New("batch").Parse(batchScript)
	if err != nil {
		return err
	}
	parameter := batchData{
		ProgramName:    filepath.Base(os.Args[0]),
		DeprecatedName: "OLD" + filepath.Base(os.Args[0]),
		UpdateFileName: updateFileName,
		BatchFileName:  batchFileName,
	}
	return batchTemplate.Execute(file, parameter)
}

func runWindowsBatch(batchFile string) error {
	cmd := exec.Command("cmd", "/c", batchFile)
	return cmd.Start()
}
