package core

import (
	"errors"
	"io/ioutil"
	"strings"

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
		return errors.New("no filename specified for ssh key")
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// publicKeyStr := string(data)
	// log.Debug(publicKeyStr)

	s.PublicKey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(data))
	if err != nil {
		panic(err)
	}

	return nil
}

// GetPublicKey will return the string form of the public key
func (s *SSHKey) GetPublicKey() string {

	original := string(ssh.MarshalAuthorizedKey(s.PublicKey))
	//  marshal adds a '\n' which linod does not like, remove it
	original = strings.TrimSuffix(original, "\n")

	// log.Debug(original)

	return original
}

// Fingerprint for the given ssh key
func (s *SSHKey) Fingerprint() string {

	fingerprint := ssh.FingerprintLegacyMD5(s.PublicKey)
	// log.Debug(fingerprint)

	return fingerprint
}
