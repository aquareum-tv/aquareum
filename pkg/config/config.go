package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/exp/rand"
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
	TLSCertPath      string
	TLSKeyPath       string
	SigningKeyPath   string
	DBPath           string
	Insecure         bool
	HttpAddr         string
	HttpsAddr        string
	HttpInternalAddr string
	AdminSecret      string
	Build            *BuildFlags
	MistAdminPort    int
	MistRTMPPort     int
	MistHTTPPort     int
}

func (cli *CLI) OwnInternalURL() string {
	//  No errors because we know it's valid from AddrFlag
	host, port, _ := net.SplitHostPort(cli.HttpInternalAddr)
	ip := net.ParseIP(host)
	if ip.IsUnspecified() {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	return fmt.Sprintf("http://%s", addr)
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

func RandomTrailer(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	res := make([]byte, length)
	for i := 0; i < length; i++ {
		res[i] = charset[rand.Intn(len(charset))]
	}
	return string(res)
}
