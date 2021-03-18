package updater

import (
	"github.com/jedisct1/go-minisign"
	"io/ioutil"
)

var UpdateFilesPubKey string

func isSignatureValid(fileName string, signatureFile string) (sigValid bool, err error) {
	pub, err := minisign.NewPublicKey(UpdateFilesPubKey)
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
