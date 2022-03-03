package tools

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"log"
	"runtime"
)

type AES_CBC struct {
	Key string //16位
	Iv  string //16位
}

func (a AES_CBC) Encrypt(value string) (string, error) {
	b, err := a.encrypt(value)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (a AES_CBC) EncryptHex(value string) (string, error) {
	b, err := a.encrypt(value)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (a AES_CBC) encrypt(value string) ([]byte, error) {
	defer Catch()
	origData := []byte(value)
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return nil, err
	}
	origData = a.PKCS5Padding(origData, block.BlockSize())
	// origData = a.ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(a.Iv))
	crypted := make([]byte, len(origData))

	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (a AES_CBC) Decrypt(crypted string) (string, error) {
	decodeData, err := base64.StdEncoding.DecodeString(crypted)
	return a.decrypt(decodeData, err)
}

func (a AES_CBC) DecryptHex(crypted string) (string, error) {
	decodeData, err := hex.DecodeString(crypted)
	return a.decrypt(decodeData, err)
}

func (a AES_CBC) decrypt(decodeData []byte, err error) (string, error) {
	defer Catch()
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
	if length == 0 {
		return origData
	}
	unpadding := int(origData[length-1])
	if length < unpadding {
		return origData
	}
	return origData[:(length - unpadding)]
}

func (a AES_CBC) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (a AES_CBC) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	if length < unpadding {
		return origData
	}
	return origData[:(length - unpadding)]
}

func (a AES_CBC) RawEncrypt(src []byte) ([]byte, error) {
	defer Catch()
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return nil, err
	}
	src = a.PKCS5Padding(src, block.BlockSize())
	// origData = a.ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(a.Iv))
	crypted := make([]byte, len(src))

	blockMode.CryptBlocks(crypted, src)
	return crypted, nil
}

func (a AES_CBC) RawDecrypt(src []byte) ([]byte, error) {
	defer Catch()
	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return nil, err
	}
	//blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, []byte(a.Iv))
	origData := make([]byte, len(src))
	blockMode.CryptBlocks(origData, src)
	origData = a.PKCS5UnPadding(origData)
	// origData = a.ZeroUnPadding(origData)
	return origData, nil
}

func Catch() {
	if r := recover(); r != nil {
		log.Println(r)
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			go log.Printf("%v,%v\n", file, line)
		}
	}
}
