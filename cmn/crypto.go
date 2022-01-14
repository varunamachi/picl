package cmn

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
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
		return nil, err
	}

	key := pbkdf2.Key([]byte(c.password), salt, 65536, 32, sha256.New)
	return key, nil
}

func (c *Cryptor) EncryptTo(reader io.Reader, writer io.Writer) error {
	return nil
}

func (c *Cryptor) DecryptFrom(reader io.Reader, writer io.Writer) error {
	return nil
}

func (c *Cryptor) Encrypt(in []byte) ([]byte, error) {
	key, err := c.getKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nil, nonce, in, nil), nil
}

func (c *Cryptor) Decrypt(in []byte) ([]byte, error) {
	key, err := c.getKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(in) < nonceSize {
		return nil, errors.New("input data is too small to decrypt")
	}

	nonce, cipherText := in[:nonceSize], in[nonceSize:]
	out, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return out, nil

	return nil, nil
}
