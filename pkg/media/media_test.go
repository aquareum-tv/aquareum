package media

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"aquareum.tv/aquareum/pkg/config"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712/eip712test"
	_ "aquareum.tv/aquareum/pkg/media/mediatesting"
	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
	"github.com/stretchr/testify/require"
)

func getFixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..", "test", "fixtures", name)
}

func mp4(t *testing.T) []byte {
	f, err := os.Open(getFixture("video.mpegts"))
	require.NoError(t, err)
	defer f.Close()
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	err = MuxToMP4(context.Background(), f, w)
	require.NoError(t, err)
	return buf.Bytes()
}

func TestMuxToMP4(t *testing.T) {
	bs := mp4(t)
	require.Greater(t, len(bs), 0)
}

func TestSignMP4(t *testing.T) {
	signer := c2pa.MakeStaticSigner(eip712test.CertBytes, eip712test.KeyBytes)
	mp4bs := mp4(t)
	r := bytes.NewReader(mp4bs)
	f, err := os.CreateTemp("", "*.mp4")
	require.NoError(t, err)
	mm := MediaManager{
		cli: &config.CLI{
			TAURL: "http://timestamp.digicert.com",
		},
		signer: signer,
		cert:   eip712test.CertBytes,
		user:   "testuser",
	}
	require.NoError(t, err)
	ms := time.Now().UnixMilli()
	err = mm.SignMP4(context.Background(), r, f, ms)
	require.NoError(t, err)
}

func TestSignMP4WithWallet(t *testing.T) {
	eip712test.WithTestSigner(func(signer *eip712.EIP712Signer) {
		certBs, err := signer.GenerateCert()
		require.NoError(t, err)
		mm := MediaManager{
			cli: &config.CLI{
				TAURL: "http://timestamp.digicert.com",
			},
			signer: signer,
			cert:   certBs,
			user:   "testuser",
		}
		mp4bs := mp4(t)
		r := bytes.NewReader(mp4bs)
		f, err := os.CreateTemp("", "*.mp4")
		require.NoError(t, err)
		ms := time.Now().UnixMilli()
		err = mm.SignMP4(context.Background(), r, f, ms)
		require.NoError(t, err)
	})
}
