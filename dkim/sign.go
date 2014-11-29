package dkim

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash"
)

type rsaSha256Signer struct {
	key *rsa.PrivateKey
}

func (this *rsaSha256Signer) Sign(data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	return this.SignHash(h)
}

func (this *rsaSha256Signer) SignHash(h hash.Hash) (string, error) {
	d := h.Sum(nil)
	result, err := rsa.SignPKCS1v15(rand.Reader, this.key, crypto.SHA256, d)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

func (this *rsaSha256Signer) signDigest(d []byte) (string, error) {
	result, err := rsa.SignPKCS1v15(rand.Reader, this.key, crypto.SHA256, d)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

func newSigner(pemBytes []byte) (*rsaSha256Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no key found")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unsupported key type %q", block.Type)
	}
	rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &rsaSha256Signer{rsa}, nil
}
