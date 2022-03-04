package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"log"
	"os"
)

type RSA struct {
	key *rsa.PrivateKey
	pub rsa.PublicKey
}

func NewRSA(bits int) RSA {
	k, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Fatal(err)
	}
	return RSA{
		key: k,
		pub: k.PublicKey,
	}
}

func NewRSAFromPKCSPrivKey(pathKey string) RSA {
	f, err := os.Open(pathKey)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	info, _ := f.Stat()
	buf := make([]byte, info.Size())
	f.Read(buf)
	blk, _ := pem.Decode(buf)
	priv, err := x509.ParsePKCS1PrivateKey(blk.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return RSA{
		key: priv,
		pub: priv.PublicKey,
	}
}

// RSA from this method only has encrypt function
func NewRSAEncryptorFromPKCSPubKey(pathPub string) RSA {
	f, err := os.Open(pathPub)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	info, _ := f.Stat()
	buf := make([]byte, info.Size())
	f.Read(buf)
	blk, _ := pem.Decode(buf)
	pki, err := x509.ParsePKIXPublicKey(blk.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	pub := pki.(*rsa.PublicKey)
	return RSA{
		pub: *pub,
	}
}

func (r *RSA) DumpPKCSKeyPairToFile() {
	X509key := x509.MarshalPKCS1PrivateKey(r.key)
	keyFile, err := os.Create("key.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer keyFile.Close()
	keyBlk := pem.Block{Type: "RSA Private Key", Bytes: X509key}
	pem.Encode(keyFile, &keyBlk)
	X509pub, err := x509.MarshalPKIXPublicKey(&r.pub)
	if err != nil {
		log.Fatal(err)
	}
	pubFile, err := os.Create("pub.pem")
	if err != nil {
		log.Fatal(err)
	}
	defer pubFile.Close()
	pubBlk := pem.Block{Type: "RSA Public Key", Bytes: X509pub}
	pem.Encode(pubFile, &pubBlk)
}

func (r *RSA) EncryptRaw(data []byte) []byte {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	enc, err := rsa.EncryptOAEP(sha256.New(), rng, &r.pub, data, label)
	if err != nil {
		log.Fatal(err)
	}
	return enc
}

func (r *RSA) DecryptRaw(data []byte) []byte {
	label := []byte("OAEP Encrypted")
	rng := rand.Reader
	dec, err := rsa.DecryptOAEP(sha256.New(), rng, r.key, data, label)
	if err != nil {
		log.Fatal(err)
	}
	return dec
}

/*

PLAINTEXT <-> RSA <-> HEX <-> CRYPTEXT

*/

func (r *RSA) EncryptStr(data string) []byte {
	return r.EncryptRaw([]byte(data))
}

func (r *RSA) EncryptToHex(data []byte) string {
	enc := r.EncryptRaw(data)
	return hex.EncodeToString(enc)
}

func (r *RSA) EncryptStrToHex(data string) string {
	enc := r.EncryptRaw([]byte(data))
	return hex.EncodeToString(enc)
}

func (r *RSA) DecryptHexStr(data string) []byte {
	edata, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	return r.DecryptRaw(edata)
}

func (r *RSA) DecryptToStr(data []byte) string {
	dec := r.DecryptRaw(data)
	return string(dec)
}

func (r *RSA) DecryptHexStrToStr(data string) string {
	edata, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	dec := r.DecryptRaw(edata)
	return string(dec)
}

func (r *RSA) PKCSEncryptRaw(data []byte) []byte {
	enc, err := rsa.EncryptPKCS1v15(rand.Reader, &r.pub, data)
	if err != nil {
		log.Fatal(err)
	}
	return enc
}

func (r *RSA) PKCSDecryptRaw(data []byte) []byte {
	dec, err := rsa.DecryptPKCS1v15(rand.Reader, r.key, data)
	if err != nil {
		log.Fatal(err)
	}
	return dec
}
