package cmn

import (
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
	ErrInput  = errors.New("invalid input")
)

const saltSize = 32
const magicSize = 4

var magic = []byte{0xE1, 0xEA, 0xE1, 0xA0}

type FileCrytor interface {
	EncryptToFile(reader io.Reader, path string) error
	DecryptFromFile(path string, writer io.Writer) error
	Encrypt(in []byte) ([]byte, error)
	Decrypt(in []byte) ([]byte, error)
	IsEncrypted(in []byte) bool
}

type aesGCMCryptor struct {
	password string
}

func NewCryptor(password string) FileCrytor {
	return &aesGCMCryptor{
		password: password,
	}
}

func (c *aesGCMCryptor) EncryptToFile(reader io.Reader, path string) error {

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

func (c *aesGCMCryptor) DecryptFromFile(path string, writer io.Writer) error {
	in, err := os.ReadFile(path)
	if err != nil {
		return Errf(err, "failed to read ciphertext from file at '%s'", path)
	}

	out, err := c.Decrypt(in)
	if err != nil {
		return err
	}

	// bw := bufio.NewWriter(writer)
	// if _, err = bw.Write(out); err != nil {
	if _, err = writer.Write(out); err != nil {
		return Errf(err, "failed write encrypted data to file")
	}
	return nil
}

func (c *aesGCMCryptor) Encrypt(in []byte) ([]byte, error) {
	if c.IsEncrypted(in) {
		return nil, Errf(ErrInput, "the data is already encrypted")
	}

	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, Errf(err, "failed to create salt")
	}

	gcm, err := c.getGCM(salt)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, Errf(err, "failed to create nonce")
	}

	val := gcm.Seal(nonce, nonce, in, nil)
	out := make([]byte, 0, len(val)+saltSize+magicSize)
	out = append(out, magic...)
	out = append(out, salt...)
	out = append(out, val...)

	return out, nil
}

func (c *aesGCMCryptor) Decrypt(in []byte) ([]byte, error) {
	if !c.IsEncrypted(in) {
		return nil, Errf(ErrInput, "the input is not properly encrypted")
	}

	in = in[magicSize:]
	salt := in[:saltSize]
	gcm, err := c.getGCM(salt)
	if err != nil {
		return nil, err
	}
	in = in[saltSize:]

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

func (c *aesGCMCryptor) getGCM(salt []byte) (cipher.AEAD, error) {

	key := pbkdf2.Key([]byte(c.password), salt, 65536, 32, sha256.New)
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

func (c *aesGCMCryptor) IsEncrypted(in []byte) bool {
	if len(in) < magicSize+saltSize {
		return false
	}
	for i := 0; i < magicSize; i++ {
		if in[i] != magic[i] {
			return false
		}
	}
	return true
}
