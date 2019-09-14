package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
)

type PrivateKey ecdsa.PrivateKey

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))

	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return (*PrivateKey)(priv), nil
}

func NewPrivateKey() (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(key), nil
}

func (p *PrivateKey) GetSerialize() string {
	x := p.X
	y := p.Y

	publicKey := x.Bytes()
	publicKey = append(publicKey, y.Bytes()...)

	dst := make([]byte, hex.DecodedLen(len(publicKey)))
	n, _ := hex.Decode(dst, publicKey)

	return string(dst[:n])
}
