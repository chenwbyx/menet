package crypto

import "bytes"

func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS5UnPadding(originalData []byte) []byte {
	length := len(originalData)
	unPadding := int(originalData[length-1])
	return originalData[:(length - unPadding)]
}
func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(0)}, padding)
	return append(cipherText, padText...)
}

func PKCS7UPadding(cipherText []byte) []byte {
	padLength := int(cipherText[len(cipherText)-1])
	return cipherText[:len(cipherText)-padLength]
}
