package vm

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

var (
	//GPG public key
	publicKey = strings.NewReader(`
	
	
	`)
	//GPG signature
	signature = strings.NewReader(`
	
	
	`)
)

//Signer to sign and verify GPG keys
type Signer struct{}

//NewSigner new instance of the signer
func NewSigner() *Signer {
	return &Signer{}
}

//Verify archives
func (signer *Signer) Verify(archivePath string) error {
	verificationTarget, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("Cannot find archive target %s", err)
	}

	keyring, err := openpgp.ReadArmoredKeyRing(publicKey)
	if err != nil {
		return fmt.Errorf("Cannot read armored key ring %s", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, verificationTarget, signature)
	if err != nil {
		return fmt.Errorf("Archive verification failed %s", err)
	}
	return nil
}
