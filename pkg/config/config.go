package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

type BuildFlags struct {
	Version   string
	BuildTime int64
	UUID      string
}

func (b BuildFlags) BuildTimeStr() string {
	ts := time.Unix(b.BuildTime, 0)
	return ts.UTC().Format(time.RFC3339)
}

func (b BuildFlags) BuildTimeStrExpo() string {
	ts := time.Unix(b.BuildTime, 0)
	return ts.UTC().Format("2006-01-02T15:04:05.000Z")
}

type CLI struct {
	TLSCertPath    string
	TLSKeyPath     string
	SigningKeyPath string
	DBPath         string
	Insecure       bool
	HttpAddr       string
	HttpsAddr      string
	AdminSecret    string
	Build          *BuildFlags
}

func (cli *CLI) ParseSigningKey() (*rsa.PrivateKey, error) {
	bs, err := os.ReadFile(cli.SigningKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bs)
	if block == nil {
		return nil, fmt.Errorf("no RSA key found in signing key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}
