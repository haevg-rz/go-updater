package updater

import (
	"encoding/json"
	"fmt"
	"github.com/artdarek/go-unzip"
	"github.com/jedisct1/go-minisign"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func (asset Asset) importUpdate(updatePath string) error {
	updateFileExtension := filepath.Ext(updatePath)
	assetFile := filepath.Join(asset.TargetFolder, asset.AssetName+updateFileExtension)
	const backUpSuffix = ".old"
	backUp := fmt.Sprint(assetFile, backUpSuffix)
	if err := os.Rename(assetFile, backUp); err != nil {
		return err
	}
	updateData, err := asset.Client.readData(updatePath)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(assetFile, updateData, 0644); err != nil {
		return err
	}
	err = unzipIfCompressed(updatePath, assetFile, asset.TargetFolder)
	return nil
}

func (asset Asset) importSelfUpdate(updatePath string) (err error) {
	data, err := asset.Client.readData(updatePath)
	if err != nil {
		return
	}
	if err = ioutil.WriteFile(updateFileName, data, 0644); err != nil {
		log.Println(err)
		return
	}
	const minisigFileExtension = ".minisig"
	data, err = asset.Client.readData(fmt.Sprint(updatePath, minisigFileExtension))
	if err != nil {
		return
	}
	signatureFile := fmt.Sprint(updateFileName, minisigFileExtension)
	if err = ioutil.WriteFile(fmt.Sprint(signatureFile), data, 0644); err != nil {
		log.Println(err)
		return
	}
	sigValid, err := isSignatureValid(updateFileName, signatureFile)
	if !sigValid || (err != nil) {
		return
	}
	zipDestination, err := os.Getwd()
	if err != nil {
		return
	}
	err = unzipIfCompressed(updatePath, updateFileName, zipDestination)
	return
}

func unzipIfCompressed(updatePath string, zipSource string, zipDestination string) (err error) {
	const compressedFileExtension = ".zip"
	if fileExtension := filepath.Ext(updatePath); fileExtension == compressedFileExtension {
		uz := unzip.New(zipSource, zipDestination)
		err = uz.Extract()
	}
	return err
}

func (asset Asset) writeVersionJson(version string) (err error) {
	const versionJsonEnding = "_Version.Json"
	fileName := fmt.Sprint(asset.AssetName, versionJsonEnding)
	filePath := filepath.Join(asset.TargetFolder, fileName)
	versionJson := &struct{ Version string }{Version: version}
	content, err := json.Marshal(versionJson)
	if err != nil {
		return
	}
	return ioutil.WriteFile(filePath, content, 0644)
}

func isSignatureValid(fileName string, signatureFile string) (sigValid bool, err error) {
	const pubKeyFile = "minisign.pub"
	pub, err := minisign.NewPublicKeyFromFile(pubKeyFile)
	if err != nil {
		return
	}
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	sig, err := minisign.NewSignatureFromFile(signatureFile)
	if err != nil {
		return
	}
	return pub.Verify(file, sig)
}
