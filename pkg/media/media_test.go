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
	"aquareum.tv/aquareum/pkg/crypto/signers"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712"
	"aquareum.tv/aquareum/pkg/crypto/signers/eip712/eip712test"
	_ "aquareum.tv/aquareum/pkg/media/mediatesting"
	"git.aquareum.tv/aquareum-tv/c2pa-go/pkg/c2pa"
	"github.com/ThalesGroup/crypto11"
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
		certBs, err := signers.GenerateES256KCert(signer)
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

func TestSignMP4WithHSM(t *testing.T) {
	one := 1
	sc, err := crypto11.Configure(&crypto11.Config{
		// TokenLabel: "C2PA Signer",
		Path:       "/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so",
		Pin:        "123456",
		SlotNumber: &one,
	})
	require.NoError(t, err)

	// bs := make([]byte, 4)
	// binary.LittleEndian.PutUint32(bs, 2)
	// signer, err := sc.FindKeyPair(nil, bs)
	// if err != nil {
	// 	return err
	// }
	// if signer == nil {
	// 	return fmt.Errorf("keypair not found")
	// }
	// pub := signer.Public()
	// fmt.Println("%v", pub)
	allsigners, err := sc.FindAllKeyPairs()
	require.NoError(t, err)
	signer := allsigners[0]
	certBs, err := signers.GenerateES256KCert(signer)
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
}
