package updater

import (
	"github.com/jedisct1/go-minisign"
	"io/ioutil"
)

var UpdateFilesPubKey string

func (a Asset) isSignatureValid(fileName string, sigPath string) (sigValid bool, err error) {
	pub, err := minisign.NewPublicKey(UpdateFilesPubKey)
	if err != nil {
		return
	}
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	pSig, err := a.getSigFromCdn(sigPath)
	if err != nil {
		return
	}
	return pub.Verify(file, *pSig)
}

func (a Asset) getSigFromCdn(sigPath string) (pSig *minisign.Signature, err error) {
	data, err := a.Client.readData(sigPath)
	if err != nil {
		return nil, err
	}
	sig, err := minisign.DecodeSignature(string(data))
	if err != nil {
		return nil, err
	}
	return &sig, nil
}
