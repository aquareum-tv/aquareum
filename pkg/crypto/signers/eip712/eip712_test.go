package eip712

import (
	"context"
	"testing"
	"time"

	v0 "aquareum.tv/aquareum/pkg/schema/v0"
	"github.com/stretchr/testify/require"
)

func makeSigner(t *testing.T) *EIP712Signer {
	schema, err := v0.MakeV0Schema()
	require.NoError(t, err)
	signer, err := MakeEIP712Signer(context.Background(), &EIP712SignerOptions{
		EthKeystorePassword: "aquareumaquareum",
		EthKeystorePath:     ".",
		EthAccountAddr:      "0x295481766f43bb048aec5d71f3bf76fdacea78f2",
		Schema:              schema,
	})
	require.NoError(t, err)
	return signer
}

func TestEIP712Map(t *testing.T) {
	msg := AquareumEIP712Message{
		MsgData:   map[string]string{"foo": "bar"},
		MsgSigner: "0x295481766f43bb048aec5d71f3bf76fdacea78f2",
		MsgTime:   time.Now().UnixMilli(),
	}
	m := msg.Map()
	require.Equal(t, m["signer"], msg.MsgSigner)
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
	_, err := signer.Sign(goLive)
	require.NoError(t, err)
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
	var goLive *v0.GoLive
	signer := makeSigner(t)
	signed, err := signer.Verify([]byte(testCase))
	require.NoError(t, err)
	require.Equal(t, signed.Signer(), "0x295481766F43bb048Aec5D71f3Bf76FDaCEA78f2")
	require.Equal(t, signed.Time(), int64(1722373018292))
	goLive, ok := signed.Data().(*v0.GoLive)
	require.True(t, ok)
	require.Equal(t, goLive.Streamer, "@aquareum.tv")
	require.Equal(t, goLive.Title, "Let's gooooooo!")
}
