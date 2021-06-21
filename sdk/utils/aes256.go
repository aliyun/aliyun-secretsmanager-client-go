package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

const (
	IterationCount   = 1000
	KeyLength        = 32
	Aes256CbcModeKey = "001"
)

// EncryptAes256Cbc 加密data，使用aes256-cbc算法，填充方式为pkcs5,
// aes密钥通过pbkdf2-hmac-sha256算法派生
func EncryptAes256Cbc(data, secret, iv, salt []byte) ([]byte, error) {
	key := pbkdf2.Key(secret, salt, IterationCount, KeyLength, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	padded := pkcs5Padding(data, block.BlockSize())
	cbc := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(padded))
	cbc.CryptBlocks(encrypted, padded)
	return encrypted, nil
}

// DecryptAes256Cbc 解密data， 使用aes256-cbc算法，填充方式为pkcs5,
// aes密钥通过pbkdf2-hmac-sha256算法派生
func DecryptAes256Cbc(data, secret, iv, salt []byte) (string, error) {
	key := pbkdf2.Key(secret, salt, IterationCount, KeyLength, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	cbc := cipher.NewCBCDecrypter(block, iv)
	cipherText := make([]byte, len(data))
	cbc.CryptBlocks(cipherText, data)
	return string(pkcs5UnPadding(cipherText)), nil
}

func pkcs5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func pkcs5UnPadding(origData []byte) []byte {
	unpadding := int(origData[len(origData)-1])
	return origData[:(len(origData) - unpadding)]
}
