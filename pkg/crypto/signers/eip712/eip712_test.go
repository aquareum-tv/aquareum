package eip712

import (
	"testing"

	v0 "aquareum.tv/aquareum/pkg/schema/v0"
	"github.com/stretchr/testify/require"
)

func makeSigner(t *testing.T) *EIP712Signer {
	signer, err := MakeEIP712Signer(&EIP712SignerOptions{
		EthKeystorePassword: "aquareumaquareum",
		EthKeystorePath:     ".",
		EthAccountAddr:      "0x295481766f43bb048aec5d71f3bf76fdacea78f2",
		Schema:              v0.Schema{},
	})
	require.NoError(t, err)
	return signer
}

func TestCreateSigner(t *testing.T) {
	makeSigner(t)
}

func TestSignGoLive(t *testing.T) {
	signer := makeSigner(t)
	goLive := v0.GoLive{
		Streamer: "@aquareum.tv",
		Title:    "Let's gooooooo!",
	}
	blob, err := signer.Sign(goLive)
	require.NoError(t, err)
	panic(string(blob))
}

var testCase = `{
  "primaryType": "GoLive",
  "domain": { "name": "Aquareum", "version": "0.0.1" },
  "message": {
    "data": { "streamer": "@aquareum.tv", "title": "Let's gooooooo!" },
    "signer": "0x295481766F43bb048Aec5D71f3Bf76FDaCEA78f2",
    "time": 1722373018292
  },
  "signature": "0x1723aa5ffb04a6ade0acb84c5ce15c804141ac06fd4ae0a867655d1b2f9e130e1ceb659297d262281795b49c191e6f67623d538890b4454eeaa1b6c2da0668e81b"
}`

func TestVerifyGoLive(t *testing.T) {

}
