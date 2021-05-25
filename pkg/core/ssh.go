package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
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

// generate a new ssh key
func (s *SSHKey) GenerateNewKey(publicKeyPath string) error {

	// generate private key file name
	privateKeyPath := strings.TrimSuffix(publicKeyPath, ".pub")

	// make sure these files don't exist already
	_, err := os.Stat(publicKeyPath)
	if (err == nil) || !os.IsNotExist(err) {
		// either file exists or some other type of error
		return err
	}
	_, err = os.Stat(privateKeyPath)
	if (err == nil) || !os.IsNotExist(err) {
		// either file exists or some other type of error
		return err
	}

	// generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// save to disk
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return err
	}

	err = privateKeyFile.Chmod(0600)
	if err != nil {
		return err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// get public key, remove new line so we can append a comment
	// finally save to disk
	publicKeyString := ssh.MarshalAuthorizedKey(pub)
	publicKeyString = bytes.TrimSuffix(publicKeyString, []byte("\n"))
	publicKeyString = append(publicKeyString, []byte(" eezhee\n")...)
	return ioutil.WriteFile(publicKeyPath, publicKeyString, 0644)
}
