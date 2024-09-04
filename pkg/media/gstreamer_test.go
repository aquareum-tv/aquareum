package media

import (
	"context"
	"os"
	"testing"

	_ "aquareum.tv/aquareum/pkg/media/mediatesting"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAudio(t *testing.T) {
	ifile, err := os.Open(getFixture("sample-stream-audio.mkv"))
	require.NoError(t, err)
	ofile, err := os.CreateTemp("", "*.mkv")
	defer os.Remove(ofile.Name())
	require.NoError(t, err)
	err = AddOpusToMKV(context.Background(), ifile, ofile)
	require.NoError(t, err)
	ofile.Close()
	info, err := os.Stat(ofile.Name())
	require.NoError(t, err)
	require.Greater(t, info.Size(), int64(0))
}
