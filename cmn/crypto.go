package cmn

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrKey    = errors.New("failed generate key")
	ErrCipher = errors.New("failed to create cipher")
	ErrFile   = errors.New("failed read/write file")
)

type Cryptor struct {
	password string
}

func NewCryptor(password string) *Cryptor {
	return &Cryptor{
		password: password,
	}
}

func (c *Cryptor) getKey() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, Errf(err, "failed to create salt")
	}

	key := pbkdf2.Key([]byte(c.password), salt, 65536, 32, sha256.New)
	return key, nil
}

func (c *Cryptor) EncryptToFile(reader io.Reader, path string) error {

	in, err := ioutil.ReadAll(reader)
	if err != nil {
		return Errf(err, "failed to read plaintext to file at %s", path)
	}

	out, err := c.Encrypt(in)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, out, 0700); err != nil {
		return Errf(err, "failed write encrypted data to file")
	}
	return nil
}

func (c *Cryptor) DecryptFromFile(path string, writer io.Writer) error {
	in, err := os.ReadFile(path)
	if err != nil {
		return Errf(err, "failed to read ciphertext from file at '%s'", path)
	}

	out, err := c.Decrypt(in)
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(writer)
	if _, err = bw.Write(out); err != nil {
		return Errf(err, "failed write encrypted data to file")
	}
	return nil
}

func (c *Cryptor) Encrypt(in []byte) ([]byte, error) {

	gcm, err := c.getGCM()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, Errf(err, "failed to create nonce")
	}

	return gcm.Seal(nil, nonce, in, nil), nil
}

func (c *Cryptor) Decrypt(in []byte) ([]byte, error) {

	gcm, err := c.getGCM()
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(in) < nonceSize {
		return nil, Errf(err, "input data is too small to decrypt")
	}

	nonce, cipherText := in[:nonceSize], in[nonceSize:]
	out, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, Errf(err, "failed to decrypt")
	}

	return out, nil

}

func (c *Cryptor) getGCM() (cipher.AEAD, error) {
	key, err := c.getKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, Errf(err, "failed to create AES cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, Errf(err, "failed to create GCM cipher")
	}

	return gcm, nil
}
