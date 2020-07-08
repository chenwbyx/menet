package crypto

import (
	"crypto/aes"
	"errors"
	"strconv"
)

func EncryptECB(plaintext []byte, key string) ([]byte, error) {
	cipher, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	blockSize := cipher.BlockSize()

	plaintext = PKCS5Padding(plaintext, blockSize)
	if len(plaintext)%blockSize != 0 {
		return nil, errors.New("need a multiple of the block size" + strconv.Itoa(blockSize))
	}

	cipherText := make([]byte, 0)
	text := make([]byte, 16)
	for len(plaintext) > 0 {
		cipher.Encrypt(text, plaintext)
		plaintext = plaintext[blockSize:]
		cipherText = append(cipherText, text...)
	}
	return cipherText, nil
}

func DecryptECB(cipherText []byte, key string) ([]byte, error) {
	cipher, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	blockSize := cipher.BlockSize()

	if len(cipherText)%blockSize != 0 {
		return nil, errors.New("need a multiple of the block size" + strconv.Itoa(blockSize))
	}

	plaintext := make([]byte, 0)
	text := make([]byte, 16)
	for len(cipherText) > 0 {
		cipher.Decrypt(text, cipherText)
		cipherText = cipherText[blockSize:]
		plaintext = append(plaintext, text...)
	}
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}
