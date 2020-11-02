package core

import (
	"errors"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

// SSHKey has details of an ssh key
type SSHKey struct {
	Name      string
	PublicKey ssh.PublicKey
}

// LoadPublicKey the public key for an ssh key
func (s *SSHKey) LoadPublicKey(filename string) error {

	if len(filename) == 0 {
		return errors.New("no filename specified")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// publicKeyStr := string(data)
	// fmt.Println(publicKeyStr)

	s.PublicKey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(data))
	if err != nil {
		panic(err)
	}

	return nil
}

// GetPublicKey will return the string form of the public key
func (s *SSHKey) GetPublicKey() string {

	original := string(ssh.MarshalAuthorizedKey(s.PublicKey))
	// fmt.Println(original)

	return original
}

// Fingerprint for the given ssh key
func (s *SSHKey) Fingerprint() string {

	fingerprint := ssh.FingerprintLegacyMD5(s.PublicKey)
	// fmt.Println(fingerprint)

	return fingerprint
}
