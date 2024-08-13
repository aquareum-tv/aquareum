package media

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "aquareum.tv/aquareum/pkg/media/mediatesting"
	"github.com/stretchr/testify/require"
)

func getFixture(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "..", "test", "fixtures", name)
}

func TestMuxToMP4(t *testing.T) {
	f, err := os.Open(getFixture("video.mpegts"))
	require.NoError(t, err)
	defer f.Close()
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	err = MuxToMP4(context.Background(), f, w)
	require.NoError(t, err)
	require.Greater(t, len(buf.Bytes()), 0)
}
