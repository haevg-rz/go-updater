package updater

import (
	"os"
	"path/filepath"
)

func (asset Asset) applyUpdate(localUpdateFile string) (err error) {
	fileExt := filepath.Ext(localUpdateFile)
	assetFile := asset.getPathToAssetFile(fileExt)
	backUpFile := asset.getPathToAssetBackUpFile(assetFile)

	if err = os.Rename(assetFile, backUpFile); err != nil {
		return err
	}
	if err = os.Rename(localUpdateFile, assetFile); err != nil {
		return err
	}
	return nil
}
