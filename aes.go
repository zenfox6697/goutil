package goutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

type AES_CBC struct {
	Key string //16位
	Iv  string //16位
}

func (a AES_CBC) Encrypt(value string) (string, error) {
	origData := []byte(value)
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData = a.PKCS5Padding(origData, blockSize)
	// origData = a.ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(a.Iv))
	crypted := make([]byte, len(origData))

	blockMode.CryptBlocks(crypted, origData)
	return hex.EncodeToString(crypted), nil

}

func (a AES_CBC) Decrypt(crypted string) (string, error) {
	decodeData, err := hex.DecodeString(crypted)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return "", err
	}
	//blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, []byte(a.Iv))
	origData := make([]byte, len(decodeData))
	blockMode.CryptBlocks(origData, decodeData)
	origData = a.PKCS5UnPadding(origData)
	// origData = a.ZeroUnPadding(origData)
	return string(origData), nil
}

func (a AES_CBC) ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func (a AES_CBC) ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func (a AES_CBC) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (a AES_CBC) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
